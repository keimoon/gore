package gore

import (
	"sync"
	"time"
)

// Sentinel is a special Redis process that monitors other Redis instances,
// does fail-over, notifies client status of all monitored instances.
type Sentinel struct {
	servers []string
	conn    *Conn
	subs    *Subscriptions
	mutex   *sync.Mutex
	state   int
}

// NewSentinel returns new Sentinel
func NewSentinel() *Sentinel {
	return &Sentinel{
		mutex: &sync.Mutex{},
		state: connStateNotConnected,
	}
}

// AddServer adds new sentinel server. Only one sentinel server is active
// at any time. If this server fails, gore will connect to other sentinel
// servers immediately.
//
// AddServer can be called at anytime, to add new server on the fly.
// In production environment, you should always have at least 3 sentinel
// servers up and running.
func (s *Sentinel) AddServer(address string) {
	for _, server := range s.servers {
		if server == address {
			return
		}
	}
	s.servers = append(s.servers, address)
}

// Init connects to one sentinel server in the list. If it fails to connect,
// it moves to the next on the list. If all servers cannot be connected,
// Init return error.
func (s *Sentinel) Init() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state != connStateNotConnected {
		return nil
	}
	return s.connect()
}

func (s *Sentinel) connect() (err error) {
	for i, server := range s.servers {
		s.conn, err = DialTimeout(server, 5*time.Second)
		if err == nil {
			continue
		}
		s.state = connStateConnected
		s.subs = NewSubscriptions(s.conn)
		s.servers = append(s.servers[0:i], s.servers[i+1:]...)
		s.servers = append(s.servers, server)
		return nil
	}
	return
}
