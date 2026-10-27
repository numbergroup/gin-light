package encoding

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/numbergroup/gin-light/internal/msgpack/common"
	"github.com/numbergroup/gin-light/internal/msgpack/def"
)

var (
	typeTime            = reflect.TypeOf(time.Time{})
	typeBinaryMarshaler = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
)

type encoder struct {
	w   io.Writer
	buf *common.Buffer
}

// Encode writes MessagePack-encoded byte array of v to writer.
func Encode(w io.Writer, v any) error {
	e := encoder{
		w:   w,
		buf: common.GetBuffer(),
	}

	err := e.create(reflect.ValueOf(v))
	if err == nil {
		err = e.buf.Flush(e.w)
	}
	common.PutBuffer(e.buf)
	return err
}

func (e *encoder) create(rv reflect.Value) error {
	if !rv.IsValid() {
		return e.writeNil()
	}

	t := rv.Type()
	if t == typeTime {
		return e.writeTime(rv.Interface().(time.Time))
	}
	if t.Kind() != reflect.Ptr && t.Kind() != reflect.Interface {
		if t.Implements(typeBinaryMarshaler) {
			return e.writeBinaryMarshaler(rv.Interface().(encoding.BinaryMarshaler))
		}
		if reflect.PointerTo(t).Implements(typeBinaryMarshaler) {
			if !rv.CanAddr() {
				p := reflect.New(t)
				p.Elem().Set(rv)
				rv = p.Elem()
			}
			return e.writeBinaryMarshaler(rv.Addr().Interface().(encoding.BinaryMarshaler))
		}
	}

	switch rv.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		v := rv.Uint()
		return e.writeUint(v)

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		v := rv.Int()
		return e.writeInt(v)

	case reflect.Float32:
		return e.writeFloat32(rv.Float())

	case reflect.Float64:
		return e.writeFloat64(rv.Float())

	case reflect.Bool:
		return e.writeBool(rv.Bool())

	case reflect.String:
		return e.writeString(rv.String())

	case reflect.Complex64:
		return e.writeComplex64(complex64(rv.Complex()))

	case reflect.Complex128:
		return e.writeComplex128(rv.Complex())

	case reflect.Slice:
		if rv.IsNil() {
			return e.writeNil()
		}
		l := rv.Len()
		// raw format
		if e.isByteSlice(rv) {
			if err := e.writeRawLength(l); err != nil {
				return err
			}
			return e.setBytes(rv.Bytes())
		}

		// format
		if err := e.writeSliceLength(l); err != nil {
			return err
		}

		if find, err := e.writeFixedSlice(rv); err != nil {
			return err
		} else if find {
			return nil
		}

		// objects
		for i := 0; i < l; i++ {
			if err := e.create(rv.Index(i)); err != nil {
				return err
			}
		}

	case reflect.Array:
		l := rv.Len()
		// raw format
		if e.isByteSlice(rv) {
			if err := e.writeRawLength(l); err != nil {
				return err
			}
			// objects
			for i := 0; i < l; i++ {
				if err := e.setByte1Uint64(rv.Index(i).Uint()); err != nil {
					return err
				}
			}
			return nil
		}

		// format
		if err := e.writeSliceLength(l); err != nil {
			return err
		}

		// objects
		for i := 0; i < l; i++ {
			if err := e.create(rv.Index(i)); err != nil {
				return err
			}
		}

	case reflect.Map:
		if rv.IsNil() {
			return e.writeNil()
		}

		l := rv.Len()
		if err := e.writeMapLength(l); err != nil {
			return err
		}

		if find, err := e.writeFixedMap(rv); err != nil {
			return err
		} else if find {
			return nil
		}

		// key-value
		iter := rv.MapRange()
		for iter.Next() {
			if err := e.create(iter.Key()); err != nil {
				return err
			}
			if err := e.create(iter.Value()); err != nil {
				return err
			}
		}

	case reflect.Struct:
		return e.writeStruct(rv)

	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return e.writeNil()
		}

		return e.create(rv.Elem())

	case reflect.Func:
		return e.writeNil()

	default:
		return fmt.Errorf("%v is %w type", rv.Kind(), def.ErrUnsupportedType)
	}
	return nil
}

func (e *encoder) writeBinaryMarshaler(m encoding.BinaryMarshaler) error {
	bs, err := m.MarshalBinary()
	if err != nil {
		return err
	}
	if bs == nil {
		return e.writeNil()
	}
	if err := e.writeRawLength(len(bs)); err != nil {
		return err
	}
	return e.setBytes(bs)
}
