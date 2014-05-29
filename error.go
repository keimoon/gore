package gore

import (
	"errors"
)

var (
	ErrNotConnected       = errors.New("not connected")
	ErrEmptyScript        = errors.New("empty script")
	ErrType               = errors.New("type error")
	ErrConvert            = errors.New("convert error")
	ErrKeyChanged         = errors.New("key changed")
	ErrTransactionAborted = errors.New("transaction aborted")
	ErrNil                = errors.New("nil value")
	ErrWrite              = errors.New("write error")
	ErrRead               = errors.New("read error")
)
