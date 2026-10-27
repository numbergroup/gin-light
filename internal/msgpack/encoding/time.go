package encoding

import (
	"time"
)

// writeTime writes t as the payload of the msgpack timestamp extension
// wrapped in a raw (str) value instead of an ext, and the zero time as nil,
// which is what ugorji/go/codec does without WriteExt.
func (e *encoder) writeTime(t time.Time) error {
	if t.IsZero() {
		return e.writeNil()
	}

	secs := uint64(t.Unix())
	if secs>>34 == 0 {
		data := uint64(t.Nanosecond())<<34 | secs
		if data&0xffffffff00000000 == 0 {
			if err := e.writeRawLength(4); err != nil {
				return err
			}
			return e.setByte4Uint64(data)
		}
		if err := e.writeRawLength(8); err != nil {
			return err
		}
		return e.setByte8Uint64(data)
	}

	if err := e.writeRawLength(12); err != nil {
		return err
	}
	if err := e.setByte4Int(t.Nanosecond()); err != nil {
		return err
	}
	return e.setByte8Uint64(secs)
}
