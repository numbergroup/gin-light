package decoding

import (
	"fmt"
	"reflect"

	"github.com/numbergroup/gin-light/internal/msgpack/common/decodingutil"
	"github.com/numbergroup/gin-light/internal/msgpack/def"
)

func (d *decoder) asInterface(k reflect.Kind) (any, error) {
	code, err := d.readSize1()
	if err != nil {
		return nil, err
	}
	return d.asInterfaceWithCode(code, k)
}

// typeInterfaceValue is the reflect.Type of any; used to size the
// element budget when pre-allocating []any.
var typeInterfaceValue = reflect.TypeOf((*any)(nil)).Elem()

func (d *decoder) asInterfaceWithCode(code byte, k reflect.Kind) (any, error) {
	switch {
	case code == def.Nil:
		return nil, nil

	case code == def.True, code == def.False:
		v, err := d.asBoolWithCode(code, k)
		if err != nil {
			return nil, err
		}
		return v, nil

	// like ugorji/go/codec: signed formats and positive fixints decode as
	// int64, unsigned formats as uint64, floats as float64 and both raw (str)
	// and bin as []byte
	case d.isPositiveFixNum(code), d.isNegativeFixNum(code),
		code == def.Int8, code == def.Int16, code == def.Int32, code == def.Int64:
		return d.asIntWithCode(code, k)

	case code == def.Uint8, code == def.Uint16, code == def.Uint32, code == def.Uint64:
		return d.asUintWithCode(code, k)

	case code == def.Float32, code == def.Float64:
		return d.asFloat64WithCode(code, k)

	case d.isCodeString(code), d.isCodeBin(code):
		return d.asRawBytesWithCode(code, k)

	case d.isFixSlice(code), code == def.Array16, code == def.Array32:
		l, err := d.sliceLength(code, k)
		if err != nil {
			return nil, err
		}

		if err = d.enter(); err != nil {
			return nil, err
		}
		v := make([]any, 0, initialSliceCap(l, typeInterfaceValue))
		for i := 0; i < l; i++ {
			vv, err := d.asInterface(k)
			if err != nil {
				return nil, err
			}
			v = append(v, vv)
		}
		d.depth--
		return v, nil

	case d.isFixMap(code), code == def.Map16, code == def.Map32:
		l, err := d.mapLength(code, k)
		if err != nil {
			return nil, err
		}

		if err = d.enter(); err != nil {
			return nil, err
		}
		v := make(map[any]any, initialMapCap(l))
		for i := 0; i < l; i++ {
			keyCode, err := d.readSize1()
			if err != nil {
				return 0, err
			}

			if err := d.canSetAsMapKey(keyCode); err != nil {
				return nil, err
			}
			key, err := d.asMapKeyWithCode(keyCode, k)
			if err != nil {
				return nil, err
			}
			if decodingutil.IsExtCode(keyCode) {
				if err := decodingutil.CheckMapKey(key); err != nil {
					return nil, err
				}
			}
			// a repeated key decodes into the existing value
			value := reflect.ValueOf(&v).Elem()
			value = reflect.New(value.Type().Elem()).Elem()
			if old, ok := v[key]; ok {
				value.Set(reflect.ValueOf(&old).Elem())
			}
			if err := d.decode(value); err != nil {
				return nil, err
			}
			v[key] = value.Interface()
		}
		d.depth--
		return v, nil
	}

	// ext
	if code == def.Fixext4 || code == def.Fixext8 || code == def.Ext8 {
		return d.asTimeWithCode(code, k)
	}
	return nil, d.errorTemplate(code, k)
}

// asMapKeyWithCode decodes a key of an interface-keyed map. Raw values are
// returned as string, since []byte cannot be a map key.
func (d *decoder) asMapKeyWithCode(code byte, k reflect.Kind) (any, error) {
	v, err := d.asInterfaceWithCode(code, k)
	if bs, ok := v.([]byte); ok {
		return string(bs), err
	}
	return v, err
}

func (d *decoder) canSetAsMapKey(code byte) error {
	switch {
	case d.isFixSlice(code), code == def.Array16, code == def.Array32:
		return fmt.Errorf("%w. code: %x", def.ErrCanNotSetSliceAsMapKey, code)
	case d.isFixMap(code), code == def.Map16, code == def.Map32:
		return fmt.Errorf("%w. code: %x", def.ErrCanNotSetMapAsMapKey, code)
	}
	return nil
}
