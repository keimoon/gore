package gore

import (
	"errors"
)

var (
	// ErrNotConnected is returned when attempt to send command when connection is down
	ErrNotConnected       = errors.New("not connected")
	// ErrEmptyScript is returned when try to execute an empty script
	ErrEmptyScript        = errors.New("empty script")
	// ErrType is returned when convert between different reply types
	ErrType               = errors.New("type error")
	// ErrConvert is returned when convert between data types
	ErrConvert            = errors.New("convert error")
	// ErrKeyChanged is returned when transaction fails because watched keys have been changed
	ErrKeyChanged         = errors.New("key changed")
	// ErrTransactionAborted is returned when tracsaction fails because of other reasons 
	ErrTransactionAborted = errors.New("transaction aborted")
	// ErrNil is for nil reply
	ErrNil                = errors.New("nil value")
	// ErrWrite is returned when connection cannot be written
	ErrWrite              = errors.New("write error")
	// ErrRead is returned when connection cannot be read
	ErrRead               = errors.New("read error")
)
