package decoding

import (
	"encoding/binary"
	"math"
	"reflect"

	"github.com/gin-gonic/gin/internal/msgpack/def"
)

func (d *decoder) asFloat64WithCode(code byte, k reflect.Kind) (float64, error) {
	switch {
	case code == def.Float64:
		bs, err := d.readSize8()
		if err != nil {
			return 0, err
		}
		v := math.Float64frombits(binary.BigEndian.Uint64(bs))
		return v, nil

	case code == def.Float32:
		bs, err := d.readSize4()
		if err != nil {
			return 0, err
		}
		v := math.Float32frombits(binary.BigEndian.Uint32(bs))
		return float64(v), nil

	// unsigned values go through int64 like in ugorji/go/codec
	case d.isPositiveFixNum(code), d.isNegativeFixNum(code),
		code == def.Uint8, code == def.Uint16, code == def.Uint32, code == def.Uint64,
		code == def.Int8, code == def.Int16, code == def.Int32, code == def.Int64:
		v, err := d.asIntWithCode(code, k)
		if err != nil {
			return 0, err
		}
		return float64(v), nil

	}
	return 0, d.errorTemplate(code, k)
}
