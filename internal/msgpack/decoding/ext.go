package decoding

import (
	"encoding/binary"

	"github.com/gin-gonic/gin/internal/msgpack/def"
)

func (d *decoder) readIfExtType(code byte) (innerType int8, data []byte, err error) {
	switch code {
	case def.Fixext1:
		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		v, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), []byte{v}, nil

	case def.Fixext2:
		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSize2()
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil

	case def.Fixext4:
		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSize4()
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil

	case def.Fixext8:
		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSize8()
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil

	case def.Fixext16:
		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSize16()
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil

	case def.Ext8:
		bs, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		size := int(bs)

		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSizeN(size)
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil

	case def.Ext16:
		bs, err := d.readSize2()
		if err != nil {
			return 0, nil, err
		}
		size := int(binary.BigEndian.Uint16(bs))

		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSizeN(size)
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil

	case def.Ext32:
		bs, err := d.readSize4()
		if err != nil {
			return 0, nil, err
		}
		size, err := lengthFromUint32(binary.BigEndian.Uint32(bs))
		if err != nil {
			return 0, nil, err
		}

		typ, err := d.readSize1()
		if err != nil {
			return 0, nil, err
		}
		data, err = d.readSizeN(size)
		if err != nil {
			return 0, nil, err
		}
		return int8(typ), data, nil
	}

	return 0, nil, nil
}
