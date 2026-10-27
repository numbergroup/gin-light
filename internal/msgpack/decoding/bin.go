package decoding

import (
	"encoding/binary"
	"reflect"

	"github.com/gin-gonic/gin/internal/msgpack/def"
)

func (d *decoder) isCodeBin(v byte) bool {
	switch v {
	case def.Bin8, def.Bin16, def.Bin32:
		return true
	}
	return false
}

func (d *decoder) asBinWithCode(code byte, k reflect.Kind) ([]byte, error) {
	switch code {
	case def.Bin8:
		l, err := d.readSize1()
		if err != nil {
			return emptyBytes, err
		}
		// avoid common buffer reference
		return d.copySizeN(int(l))

	case def.Bin16:
		bs, err := d.readSize2()
		if err != nil {
			return emptyBytes, err
		}
		// avoid common buffer reference
		return d.copySizeN(int(binary.BigEndian.Uint16(bs)))

	case def.Bin32:
		bs, err := d.readSize4()
		if err != nil {
			return emptyBytes, err
		}
		l, err := lengthFromUint32(binary.BigEndian.Uint32(bs))
		if err != nil {
			return emptyBytes, err
		}
		// avoid common buffer reference
		return d.copySizeN(l)
	}

	return emptyBytes, d.errorTemplate(code, k)
}

// asRawBytesWithCode reads a str or bin value. Like ugorji/go/codec it
// also accepts nil as empty and an array of integers as its bytes.
func (d *decoder) asRawBytesWithCode(code byte, k reflect.Kind) ([]byte, error) {
	switch {
	case d.isCodeBin(code):
		return d.asBinWithCode(code, k)
	case d.isCodeString(code):
		return d.asStringByteWithCode(code, k)
	case d.isCodeNil(code):
		return emptyBytes, nil
	case d.isFixSlice(code), code == def.Array16, code == def.Array32:
		var bs []byte
		if err := d.decodeWithCode(code, reflect.ValueOf(&bs).Elem()); err != nil {
			return emptyBytes, err
		}
		return bs, nil
	}
	return emptyBytes, d.errorTemplate(code, k)
}

func (d *decoder) copySizeN(n int) ([]byte, error) {
	if n <= len(d.buf.Data) {
		bs, err := d.readSizeN(n)
		if err != nil {
			return emptyBytes, err
		}
		v := make([]byte, n)
		copy(v, bs)
		return v, nil
	}
	return d.readSizeN(n)
}
