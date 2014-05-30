package gore

import (
	"fmt"
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
func (cmd *Command) Run(conn *Conn) (r *Reply, err error) {
	if conn.state != connStateConnected {
		return nil, ErrNotConnected
	}
	conn.Lock()
	defer func() {
		conn.Unlock()
		if err != nil {
			conn.fail()
		}
	}()
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetWriteDeadline(time.Now().Add(conn.RequestTimeout))
	}
	err = cmd.writeCommand(conn)
	if err != nil {
		return nil, ErrWrite
	}
	err = conn.wb.Flush()
	if err != nil {
		return nil, ErrWrite
	}
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetReadDeadline(time.Now().Add(conn.RequestTimeout))
	}
	return readReply(conn)
}

// Send safely sends a command over conn
func (cmd *Command) Send(conn *Conn) (err error) {
	conn.Lock()
	defer func() {
                conn.Unlock()
                if err != nil {
                        conn.fail()
                }
	}()
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetWriteDeadline(time.Now().Add(conn.RequestTimeout))
	}
	err = cmd.writeCommand(conn)
	if err != nil {
		return ErrWrite
	}
	err = conn.wb.Flush()
	if err != nil {
		return ErrWrite
	}
	return nil
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
