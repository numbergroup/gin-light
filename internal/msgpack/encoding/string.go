package encoding

import (
	"math"
	"unsafe"

	"github.com/numbergroup/gin-light/internal/msgpack/def"
)

func (e *encoder) writeString(str string) error {
	// NOTE : unsafe
	strBytes := *(*[]byte)(unsafe.Pointer(&str))
	if err := e.writeRawLength(len(strBytes)); err != nil {
		return err
	}
	return e.setBytes(strBytes)
}

// writeRawLength writes the header of a string or byte slice. Like
// ugorji/go/codec without WriteExt, both use the raw (str) family and
// skip str8, which predates the current spec.
func (e *encoder) writeRawLength(l int) error {
	if l < 32 {
		if err := e.setByte1Int(def.FixStr + l); err != nil {
			return err
		}
	} else if l <= math.MaxUint16 {
		if err := e.setByte1Int(def.Str16); err != nil {
			return err
		}
		if err := e.setByte2Int(l); err != nil {
			return err
		}
	} else {
		if err := e.setByte1Int(def.Str32); err != nil {
			return err
		}
		if err := e.setByte4Int(l); err != nil {
			return err
		}
	}
	return nil
}
