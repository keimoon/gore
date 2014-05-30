package gore

import (
	"container/list"
	"sync"
	"time"
)

// Pool is a pool of connection. The application acquires connection
// from pool using Acquire() method, and when done, returns it to the pool
// with Release().
type Pool struct {
	l       *list.List
	mutex   sync.Mutex
	address string
	// Request timeout for each connection
	RequestTimeout time.Duration
	// Initial number of connection to open
	InitialConn int
	// Maximum number of connection to open
	MaximumConn int
}
