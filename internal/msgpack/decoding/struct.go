package decoding

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/numbergroup/gin-light/internal/msgpack/common"
	"github.com/numbergroup/gin-light/internal/msgpack/common/decodingutil"
	"github.com/numbergroup/gin-light/internal/msgpack/def"
)

type structInfo struct {
	fields []common.Field
	byName map[string]int
}

var structInfoCache sync.Map // map[reflect.Type]*structInfo

func getStructInfo(t reflect.Type) *structInfo {
	if si, ok := structInfoCache.Load(t); ok {
		return si.(*structInfo)
	}
	fields := common.Fields(t)
	si := &structInfo{fields: fields, byName: make(map[string]int, len(fields))}
	for i, f := range fields {
		si.byName[f.Name] = i
	}
	v, _ := structInfoCache.LoadOrStore(t, si)
	return v.(*structInfo)
}

func (d *decoder) setStruct(code byte, rv reflect.Value, k reflect.Kind) error {
	if err := d.enter(); err != nil {
		return err
	}
	var err error
	if d.isFixSlice(code) || code == def.Array16 || code == def.Array32 {
		err = d.setStructFromArray(code, rv, k)
	} else {
		err = d.setStructFromMap(code, rv, k)
	}
	if err != nil {
		return err
	}
	d.depth--
	return nil
}

func (d *decoder) setStructFromArray(code byte, rv reflect.Value, k reflect.Kind) error {
	// get length
	l, err := d.sliceLength(code, k)
	if err != nil {
		return err
	}

	si := getStructInfo(rv.Type())
	// set value
	for i := 0; i < l; i++ {
		if i < len(si.fields) {
			err = d.decodeField(rv, si.fields[i].Index)
			if err != nil {
				return err
			}
		} else {
			err = d.jumpOffset()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *decoder) setStructFromMap(code byte, rv reflect.Value, k reflect.Kind) error {
	if !d.isFixMap(code) && code != def.Map16 && code != def.Map32 {
		return fmt.Errorf("%w: only encoded map or array can decode into struct, got %x", def.ErrCanNotDecode, code)
	}
	// get length
	l, err := d.mapLength(code, k)
	if err != nil {
		return err
	}

	si := getStructInfo(rv.Type())
	for i := 0; i < l; i++ {
		keyCode, err := d.readSize1()
		if err != nil {
			return err
		}
		dataKey, err := d.asRawBytesWithCode(keyCode, k)
		if err != nil {
			return err
		}

		if fi, ok := si.byName[string(dataKey)]; ok {
			err = d.decodeField(rv, si.fields[fi].Index)
			if err != nil {
				return err
			}
		} else {
			err = d.jumpOffset()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// decodeField decodes the next value into the field of rv at index. Nil
// embedded pointers on the way are allocated, unless the value is nil.
func (d *decoder) decodeField(rv reflect.Value, index []int) error {
	code, err := d.readSize1()
	if err != nil {
		return err
	}
	f, err := fieldByIndex(rv, index, !d.isCodeNil(code))
	if err != nil || !f.IsValid() {
		return err
	}
	return d.decodeWithCode(code, f)
}

// fieldByIndex is like reflect.Value.FieldByIndex but allocates nil
// embedded pointers on the way, or returns an invalid value if alloc is
// false.
func fieldByIndex(rv reflect.Value, index []int, alloc bool) (reflect.Value, error) {
	for i, x := range index {
		if i > 0 && rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				if !alloc {
					return reflect.Value{}, nil
				}
				if !rv.CanSet() {
					return reflect.Value{}, fmt.Errorf("%w: cannot set embedded pointer to unexported struct %v", def.ErrUnsupportedType, rv.Type().Elem())
				}
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		rv = rv.Field(x)
	}
	return rv, nil
}

func (d *decoder) jumpOffset() error {
	code, err := d.readSize1()
	if err != nil {
		return err
	}

	switch {
	case code == def.True, code == def.False, code == def.Nil:
		// do nothing

	case d.isPositiveFixNum(code) || d.isNegativeFixNum(code):
		// do nothing
	case code == def.Uint8, code == def.Int8:
		_, err = d.readSize1()
		return err
	case code == def.Uint16, code == def.Int16:
		_, err = d.readSize2()
		return err
	case code == def.Uint32, code == def.Int32, code == def.Float32:
		_, err = d.readSize4()
		return err
	case code == def.Uint64, code == def.Int64, code == def.Float64:
		_, err = d.readSize8()
		return err

	case d.isFixString(code):
		_, err = d.readSizeN(int(code - def.FixStr))
		return err
	case code == def.Str8, code == def.Bin8:
		b, err := d.readSize1()
		if err != nil {
			return err
		}
		_, err = d.readSizeN(int(b))
		return err
	case code == def.Str16, code == def.Bin16:
		bs, err := d.readSize2()
		if err != nil {
			return err
		}
		_, err = d.readSizeN(int(binary.BigEndian.Uint16(bs)))
		return err
	case code == def.Str32, code == def.Bin32:
		bs, err := d.readSize4()
		if err != nil {
			return err
		}
		l, err := lengthFromUint32(binary.BigEndian.Uint32(bs))
		if err != nil {
			return err
		}
		_, err = d.readSizeN(l)
		return err

	case d.isFixSlice(code):
		l := int(code - def.FixArray)
		for i := 0; i < l; i++ {
			if err = d.jumpOffset(); err != nil {
				return err
			}
		}
	case code == def.Array16:
		bs, err := d.readSize2()
		if err != nil {
			return err
		}
		l := int(binary.BigEndian.Uint16(bs))
		for i := 0; i < l; i++ {
			if err = d.jumpOffset(); err != nil {
				return err
			}
		}
	case code == def.Array32:
		bs, err := d.readSize4()
		if err != nil {
			return err
		}
		l, err := lengthFromUint32(binary.BigEndian.Uint32(bs))
		if err != nil {
			return err
		}
		for i := 0; i < l; i++ {
			if err = d.jumpOffset(); err != nil {
				return err
			}
		}

	case d.isFixMap(code):
		l := int(code - def.FixMap)
		for i := 0; i < l*2; i++ {
			if err = d.jumpOffset(); err != nil {
				return err
			}
		}
	case code == def.Map16:
		bs, err := d.readSize2()
		if err != nil {
			return err
		}
		l := int(binary.BigEndian.Uint16(bs))
		for i := 0; i < l*2; i++ {
			if err = d.jumpOffset(); err != nil {
				return err
			}
		}
	case code == def.Map32:
		bs, err := d.readSize4()
		if err != nil {
			return err
		}
		l, err := lengthFromUint32(binary.BigEndian.Uint32(bs))
		if err != nil {
			return err
		}
		// skip key and value separately: multiplying the pair count would
		// overflow on 32-bit platforms
		for i := 0; i < l; i++ {
			if err = d.jumpOffset(); err != nil {
				return err
			}
			if err = d.jumpOffset(); err != nil {
				return err
			}
		}

	case code == def.Fixext1:
		_, err = d.readSizeN(def.Byte1 + def.Byte1)
		return err
	case code == def.Fixext2:
		_, err = d.readSizeN(def.Byte1 + def.Byte2)
		return err
	case code == def.Fixext4:
		_, err = d.readSizeN(def.Byte1 + def.Byte4)
		return err
	case code == def.Fixext8:
		_, err = d.readSizeN(def.Byte1 + def.Byte8)
		return err
	case code == def.Fixext16:
		_, err = d.readSizeN(def.Byte1 + def.Byte16)
		return err

	case code == def.Ext8:
		b, err := d.readSize1()
		if err != nil {
			return err
		}
		_, err = d.readSizeN(def.Byte1 + int(b))
		return err
	case code == def.Ext16:
		bs, err := d.readSize2()
		if err != nil {
			return err
		}
		_, err = d.readSizeN(def.Byte1 + int(binary.BigEndian.Uint16(bs)))
		return err
	case code == def.Ext32:
		bs, err := d.readSize4()
		if err != nil {
			return err
		}
		l, err := lengthFromUint32(binary.BigEndian.Uint32(bs))
		if err != nil {
			return err
		}
		if l >= math.MaxInt { // Byte1 + l must not overflow
			return decodingutil.ErrDeclaredLengthTooLarge
		}
		_, err = d.readSizeN(def.Byte1 + l)
		return err

	default:
		return d.errorTemplate(code, reflect.Invalid)
	}
	return nil
}
