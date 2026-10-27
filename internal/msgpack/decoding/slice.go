package decoding

import (
	"encoding/binary"
	"math"
	"reflect"

	"github.com/gin-gonic/gin/internal/msgpack/common/decodingutil"
	"github.com/gin-gonic/gin/internal/msgpack/def"
)

func (d *decoder) isFixSlice(v byte) bool {
	return def.FixArray <= v && v <= def.FixArray+0x0f
}

func (d *decoder) isCodeSlice(v byte) bool {
	return d.isFixSlice(v) || v == def.Array16 || v == def.Array32
}

// seqLength returns the number of elements to decode into a slice or array.
// Like ugorji/go/codec, a map is read as its keys and values in sequence.
func (d *decoder) seqLength(code byte, k reflect.Kind) (int, error) {
	if !d.isFixMap(code) && code != def.Map16 && code != def.Map32 {
		return d.sliceLength(code, k)
	}
	l, err := d.mapLength(code, k)
	if err != nil {
		return 0, err
	}
	if l > math.MaxInt/2 {
		return 0, decodingutil.ErrDeclaredLengthTooLarge
	}
	return l * 2, nil
}

func (d *decoder) sliceLength(code byte, k reflect.Kind) (int, error) {
	switch {
	case d.isFixSlice(code):
		return int(code - def.FixArray), nil
	case code == def.Array16:
		bs, err := d.readSize2()
		if err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint16(bs)), nil
	case code == def.Array32:
		bs, err := d.readSize4()
		if err != nil {
			return 0, err
		}
		return lengthFromUint32(binary.BigEndian.Uint32(bs))
	}
	return 0, d.errorTemplate(code, k)
}
