package decoding

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/numbergroup/gin-light/internal/msgpack/def"
)

// asTimeWithCode decodes a time from either the msgpack timestamp extension
// or, as written by ugorji/go/codec without WriteExt, the same payload
// wrapped in a raw (str) or bin value.
func (d *decoder) asTimeWithCode(code byte, k reflect.Kind) (time.Time, error) {
	if d.isCodeString(code) || d.isCodeBin(code) {
		bs, err := d.asRawBytesWithCode(code, k)
		if err != nil {
			return time.Time{}, err
		}
		return decodeTimestamp(bs)
	}

	typ, data, err := d.readIfExtType(code)
	if err != nil {
		return time.Time{}, err
	}
	if data == nil {
		return time.Time{}, d.errorTemplate(code, k)
	}
	if typ != def.TimeStamp {
		return time.Time{}, fmt.Errorf("%w: ext type %d decoding as time", def.ErrCanNotDecode, typ)
	}
	return decodeTimestamp(data)
}

func decodeTimestamp(bs []byte) (time.Time, error) {
	switch len(bs) {
	case 4:
		return time.Unix(int64(binary.BigEndian.Uint32(bs)), 0).UTC(), nil

	case 8:
		data64 := binary.BigEndian.Uint64(bs)
		nano := int64(data64 >> 34)
		return time.Unix(int64(data64&0x00000003ffffffff), nano).UTC(), nil

	case 12:
		nano := binary.BigEndian.Uint32(bs[:4])
		sec := binary.BigEndian.Uint64(bs[4:12])
		return time.Unix(int64(sec), int64(nano)).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("%w: invalid length of bytes for decoding time - expecting 4 or 8 or 12, got %d", def.ErrCanNotDecode, len(bs))
}
