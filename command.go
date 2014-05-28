package gore

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Command sent to redis
type Command struct {
	name string
	args []interface{}
}

// NewCommand returns a new Command
func NewCommand(name string, args ...interface{}) *Command {
	return &Command{
		name: strings.TrimSpace(name),
		args: args,
	}
}

// Run sends command to redis
func (cmd *Command) Run(conn *Conn) (*Reply, error) {
	if cmd.name == "" {
		return nil, ErrCommandEmpty
	}
	conn.mutex.Lock()
	if conn.state != connStateConnected {
		conn.mutex.Unlock()
		return nil, ErrNotConnected
	}
	conn.mutex.Unlock()
	conn.writeMutex.Lock()
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetWriteDeadline(time.Now().Add(conn.RequestTimeout))
	}
	err := cmd.writeCommand(conn)
	if err != nil {
		conn.writeMutex.Unlock()
		conn.reconnect()
		return nil, ErrWrite
	}
	err = conn.wb.Flush()
	if err != nil {
                conn.writeMutex.Unlock()
                conn.reconnect()
                return nil, ErrWrite
        }
	// Djiskstra will not like this
	conn.readMutex.Lock()
	conn.writeMutex.Unlock()
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetReadDeadline(time.Now().Add(conn.RequestTimeout))
	}
	rep, err := readReply(conn)
	conn.readMutex.Unlock()
	if err != nil {
		conn.reconnect()
	}
	return rep, err
}

func (cmd *Command) writeCommand(conn *Conn) error {
	cmdLen := strconv.FormatInt(int64(len(cmd.args))+1, 10)
	_, err := conn.wb.WriteString("*" + cmdLen + "\r\n")
	if err != nil {
		return err
	}
	err = writeString(cmd.name, conn)
	if err != nil {
		return err
	}
	for _, arg := range cmd.args {
		switch arg := arg.(type) {
		case string:
			err = writeString(arg, conn)
		case []byte:
			err = writeBytes(arg, conn)
		case int:
			err = writeString(strconv.FormatInt(int64(arg), 10), conn)
		case int64:
			err = writeString(strconv.FormatInt(arg, 10), conn)
		case float64:
			err = writeString(strconv.FormatFloat(arg, 'g', -1, 64), conn)
		case FixInt:
			err = writeBytes(arg.Bytes(), conn)
		case VarInt:
			err = writeBytes(arg.Bytes(), conn)
		case bool:
			if arg {
				err = writeString("1", conn)
			} else {
				err = writeString("0", conn)
			}
		case nil:
			err = writeString("", conn)
		default:
			err = writeString(fmt.Sprint(arg), conn)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func writeString(s string, conn *Conn) error {
	l := strconv.FormatInt(int64(len(s)), 10)
	_, err := conn.wb.WriteString("$" + l + "\r\n" + s + "\r\n")
	return err
}

func writeBytes(b []byte, conn *Conn) error {
	l := strconv.FormatInt(int64(len(b)), 10)
	conn.wb.WriteString("$" + l + "\r\n")
	conn.wb.Write(b)
	_, err := conn.wb.WriteString("\r\n")
	return err
}

// Motivated by redigo. Good job, man
func readReply(conn *Conn) (*Reply, error) {
	line, err := readLine(conn)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, ErrRead
	}
	switch line[0] {
	case '+':
		switch {
		case len(line) == 3 && line[1] == 'O' && line[2] == 'K':
			return okReply, nil
		case len(line) == 5 && line[1] == 'P' && line[2] == 'O' && line[3] == 'N' && line[4] == 'G':
			return pongReply, nil
		default:
			return &Reply{
				replyType:   ReplyStatus,
				stringValue: line[1:],
			}, nil

		}
	case '-':
		return &Reply{
			replyType:   ReplyError,
			stringValue: line[1:],
		}, nil
	case ':':
		intValue, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, ErrRead
		}
		return &Reply{
			replyType:    ReplyInteger,
			integerValue: intValue,
		}, nil
	case '$':
		l, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, ErrRead
		}
		if l < 0 {
			return &Reply{
				replyType: ReplyNil,
			}, nil
		}
		b := make([]byte, l)
		_, err = io.ReadFull(conn.rb, b)
		if err != nil {
			return nil, ErrRead
		}
		line, err = readLine(conn)
		if err != nil || len(line) != 0 {
			return nil, ErrRead
		}
		return &Reply{
			replyType:   ReplyString,
			stringValue: b,
		}, nil
	case '*':
		l, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, ErrRead
		}
		if l < 0 {
			return &Reply{
				replyType: ReplyNil,
			}, nil
		}
		replyArray := make([]*Reply, l)
		for i := range replyArray {
			replyArray[i], err = readReply(conn)
			if err != nil {
				return nil, err
			}
		}
		return &Reply{
			replyType:  ReplyArray,
			arrayValue: replyArray,
		}, nil
	default:
		return nil, ErrRead
	}
}

func readLine(conn *Conn) ([]byte, error) {
	b, err := conn.rb.ReadSlice('\n')
	if err != nil {
		return nil, ErrRead
	}
	i := len(b) - 2
	if i < 0 || b[i] != '\r' {
		return nil, ErrRead
	}
	return b[:i], nil
}
