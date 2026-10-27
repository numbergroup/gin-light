package decoding

import (
	"encoding/binary"
	"reflect"

	"github.com/numbergroup/gin-light/internal/msgpack/def"
)

func (d *decoder) isFixMap(v byte) bool {
	return def.FixMap <= v && v <= def.FixMap+0x0f
}

func (d *decoder) mapLength(code byte, k reflect.Kind) (int, error) {
	switch {
	case d.isFixMap(code):
		return int(code - def.FixMap), nil
	case code == def.Map16:
		bs, err := d.readSize2()
		if err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint16(bs)), nil
	case code == def.Map32:
		bs, err := d.readSize4()
		if err != nil {
			return 0, err
		}
		return lengthFromUint32(binary.BigEndian.Uint32(bs))
	}

	return 0, d.errorTemplate(code, k)
}
