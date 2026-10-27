package decoding

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin/internal/msgpack/def"
)

func (d *decoder) asUintWithCode(code byte, k reflect.Kind) (uint64, error) {
	switch {
	case d.isPositiveFixNum(code):
		return uint64(code), nil

	case code == def.Uint8:
		b, err := d.readSize1()
		if err != nil {
			return 0, err
		}
		return uint64(b), nil

	case code == def.Uint16:
		bs, err := d.readSize2()
		if err != nil {
			return 0, err
		}
		v := binary.BigEndian.Uint16(bs)
		return uint64(v), nil

	case code == def.Uint32:
		bs, err := d.readSize4()
		if err != nil {
			return 0, err
		}
		v := binary.BigEndian.Uint32(bs)
		return uint64(v), nil

	case code == def.Uint64:
		bs, err := d.readSize8()
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint64(bs), nil

	case d.isNegativeFixNum(code), code == def.Int8, code == def.Int16, code == def.Int32, code == def.Int64:
		v, err := d.asIntWithCode(code, k)
		if err != nil {
			return 0, err
		}
		if v < 0 {
			return 0, fmt.Errorf("%w: assigning negative signed value %d to unsigned type", def.ErrCanNotDecode, v)
		}
		return uint64(v), nil

	case code == def.Float32, code == def.Float64:
		f, err := d.asFloat64WithCode(code, k)
		if err != nil {
			return 0, err
		}
		v := uint64(f)
		if f < 0 || float64(v) != f {
			return 0, fmt.Errorf("%w: assigning unsigned integer value from float64 %v", def.ErrCanNotDecode, f)
		}
		return v, nil
	}

	return 0, d.errorTemplate(code, k)
}
