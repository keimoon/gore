package gore

import (
	"time"
)

// Pipeline keeps a list of command for sending to redis once, saving network roundtrip
type Pipeline struct {
	commands []*Command
}

// NewPipeline returns new Pipeline
func NewPipeline() *Pipeline {
	return &Pipeline{
		commands: []*Command{},
	}
}

// Add appends new commands to the pipeline
func (p *Pipeline) Add(cmd ...*Command) {
	p.commands = append(p.commands, cmd...)
}

// Reset clears all command in the pipeline
func (p *Pipeline) Reset() {
	p.commands = []*Command{}
}

// Run sends the pipeline and returns a slice of Reply
func (p *Pipeline) Run(conn *Conn) ([]*Reply, error) {
	if len(p.commands) == 0 {
		return nil, nil
	}
	conn.Lock()
	if conn.state != connStateConnected {
		conn.Unlock()
		return nil, ErrNotConnected
	}
	conn.Unlock()
	conn.LockWrite()
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetWriteDeadline(time.Now().Add(conn.RequestTimeout * time.Duration(len(p.commands) / 10 + 1)))
	}
	for _, cmd := range p.commands {
		err := cmd.writeCommand(conn)
		if err != nil {
			conn.UnlockWrite()
			conn.fail()
			return nil, ErrWrite
		}
	}
	err := conn.wb.Flush()
	if err != nil {
		conn.UnlockWrite()
		conn.fail()
		return nil, ErrWrite
	}
	conn.LockRead()
	conn.UnlockWrite()
	if conn.RequestTimeout != 0 {
		conn.tcpConn.SetReadDeadline(time.Now().Add(conn.RequestTimeout * time.Duration(len(p.commands) / 10 + 1)))
	}
	replies := make([]*Reply, len(p.commands))
	for i := range replies {
		rep, err := readReply(conn)
		if err != nil {
			conn.UnlockRead()
			conn.fail()
			return nil, err
		}
		replies[i] = rep
	}
	conn.UnlockRead()
	return replies, nil
}
