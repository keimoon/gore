package gore

import (
	"errors"
)

var (
	ErrNotConnected = errors.New("not connected")
	ErrCommandEmpty = errors.New("empty command")
	ErrType         = errors.New("type error")
	ErrConvert      = errors.New("convert error")
	ErrNil          = errors.New("nil value")
	ErrWrite        = errors.New("write error")
	ErrRead         = errors.New("read error")
)
