package def

import (
	"errors"
	"fmt"
)

var (
	// base errors

	ErrMsgpack = errors.New("msgpack: ")

	// decoding errors

	ErrNoData                 = fmt.Errorf("%wno data", ErrMsgpack)
	ErrReceiverNotPointer     = fmt.Errorf("%wreceiver not pointer", ErrMsgpack)
	ErrNotMatchArrayElement   = fmt.Errorf("%wnot match array element", ErrMsgpack)
	ErrCanNotDecode           = fmt.Errorf("%winvalid code", ErrMsgpack)
	ErrCanNotSetSliceAsMapKey = fmt.Errorf("%wcan not set slice as map key", ErrMsgpack)
	ErrCanNotSetMapAsMapKey   = fmt.Errorf("%wcan not set map as map key", ErrMsgpack)
	ErrUnhashableMapKey       = fmt.Errorf("%wunhashable map key", ErrMsgpack)
	ErrExceededMaxDepth       = fmt.Errorf("%wexceeded max nesting depth", ErrMsgpack)

	// encoding errors

	ErrUnsupportedType = fmt.Errorf("%wunsupported type", ErrMsgpack)
)
