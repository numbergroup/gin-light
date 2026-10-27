package decoding

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/gin-gonic/gin/internal/msgpack/common"
	"github.com/gin-gonic/gin/internal/msgpack/common/decodingutil"
	"github.com/gin-gonic/gin/internal/msgpack/def"
)

var (
	typeTime              = reflect.TypeOf(time.Time{})
	typeBytes             = reflect.TypeOf([]byte(nil))
	typeBinaryUnmarshaler = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

type decoder struct {
	r     io.Reader
	buf   *common.Buffer
	depth int32
}

// Decode analyzes the MessagePack-encoded data and stores
// the result into the pointer of v.
func Decode(r io.Reader, v any) error {
	if r == nil {
		return def.ErrNoData
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%w. v.(type): %T", def.ErrReceiverNotPointer, v)
	}

	rv = rv.Elem()

	d := decoder{
		r:   r,
		buf: common.GetBuffer(),
	}
	err := d.decode(rv)
	common.PutBuffer(d.buf)
	return err
}

func (d *decoder) decode(rv reflect.Value) error {
	code, err := d.readSize1()
	if err != nil {
		return err
	}
	return d.decodeWithCode(code, rv)
}

func (d *decoder) decodeWithCode(code byte, rv reflect.Value) error {
	// nil resets the receiver to its zero value
	if d.isCodeNil(code) {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}

	t := rv.Type()
	if t == typeTime {
		v, err := d.asTimeWithCode(code, rv.Kind())
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(v))
		return nil
	}
	k := rv.Kind()
	if k != reflect.Ptr && k != reflect.Interface && rv.CanAddr() && reflect.PointerTo(t).Implements(typeBinaryUnmarshaler) {
		bs, err := d.asRawBytesWithCode(code, k)
		if err != nil {
			return err
		}
		return rv.Addr().Interface().(encoding.BinaryUnmarshaler).UnmarshalBinary(bs)
	}

	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := d.asIntWithCode(code, k)
		if err != nil {
			return err
		}
		if rv.OverflowInt(v) {
			return fmt.Errorf("%w: %d overflows %v", def.ErrCanNotDecode, v, t)
		}
		rv.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := d.asUintWithCode(code, k)
		if err != nil {
			return err
		}
		if rv.OverflowUint(v) {
			return fmt.Errorf("%w: %d overflows %v", def.ErrCanNotDecode, v, t)
		}
		rv.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := d.asFloat64WithCode(code, k)
		if err != nil {
			return err
		}
		if rv.OverflowFloat(v) {
			return fmt.Errorf("%w: %g overflows %v", def.ErrCanNotDecode, v, t)
		}
		rv.SetFloat(v)

	case reflect.Complex64, reflect.Complex128:
		v, err := d.asFloat64WithCode(code, k)
		if err != nil {
			return err
		}
		rv.SetComplex(complex(v, 0))

	case reflect.String:
		bs, err := d.asRawBytesWithCode(code, k)
		if err != nil {
			return err
		}
		rv.SetString(string(bs))

	case reflect.Bool:
		v, err := d.asBoolWithCode(code, k)
		if err != nil {
			return err
		}
		rv.SetBool(v)

	case reflect.Slice:
		// byte slice from raw (str) or bin
		if t.Elem().Kind() == reflect.Uint8 && (d.isCodeString(code) || d.isCodeBin(code)) {
			bs, err := d.asRawBytesWithCode(code, k)
			if err != nil {
				return err
			}
			rv.SetBytes(bs)
			return nil
		}
		if t == typeBytes && !d.isCodeSlice(code) {
			return d.errorTemplate(code, k)
		}

		// get slice length
		l, err := d.seqLength(code, k)
		if err != nil {
			return err
		}

		if err = d.enter(); err != nil {
			return err
		}
		tmpSlice := reflect.MakeSlice(t, 0, initialSliceCap(l, t.Elem()))
		for i := 0; i < l; i++ {
			tmpSlice = reflect.Append(tmpSlice, reflect.Zero(t.Elem()))
			if err = d.decode(tmpSlice.Index(i)); err != nil {
				return err
			}
		}
		rv.Set(tmpSlice)
		d.depth--

	case reflect.Array:
		// byte array from raw (str) or bin
		if t.Elem().Kind() == reflect.Uint8 && (d.isCodeString(code) || d.isCodeBin(code)) {
			bs, err := d.asRawBytesWithCode(code, k)
			if err != nil {
				return err
			}
			if len(bs) > rv.Len() {
				return fmt.Errorf("%v len is %d, but msgpack has %d elements, %w", t, rv.Len(), len(bs), def.ErrNotMatchArrayElement)
			}
			for i, b := range bs {
				rv.Index(i).SetUint(uint64(b))
			}
			return nil
		}

		// get slice length
		l, err := d.seqLength(code, k)
		if err != nil {
			return err
		}

		// elements beyond the array length are skipped
		if err = d.enter(); err != nil {
			return err
		}
		for i := 0; i < l; i++ {
			if i < rv.Len() {
				err = d.decode(rv.Index(i))
			} else {
				err = d.jumpOffset()
			}
			if err != nil {
				return err
			}
		}
		d.depth--

	case reflect.Map:
		// get map length
		l, err := d.mapLength(code, k)
		if err != nil {
			return err
		}

		// decoding into an existing map merges into it
		if err = d.enter(); err != nil {
			return err
		}
		key := t.Key()
		value := t.Elem()
		ifaceKey := key.Kind() == reflect.Interface && key.NumMethod() == 0
		checkKey := decodingutil.MapKeyNeedsCheck(key)
		if rv.IsNil() {
			rv.Set(reflect.MakeMapWithSize(t, initialMapCapForType(l, key, value)))
		}
		for i := 0; i < l; i++ {
			k := reflect.New(key).Elem()
			v := reflect.New(value).Elem()
			var keyCode byte
			keyCode, err = d.readSize1()
			if err != nil {
				return err
			}
			extKey := false
			if ifaceKey {
				if err = d.canSetAsMapKey(keyCode); err != nil {
					return err
				}
				extKey = decodingutil.IsExtCode(keyCode)
				var kv any
				kv, err = d.asMapKeyWithCode(keyCode, k.Kind())
				if kv != nil {
					k.Set(reflect.ValueOf(kv))
				}
			} else {
				err = d.decodeWithCode(keyCode, k)
			}
			if err != nil {
				return err
			}
			// a repeated key decodes into the existing value
			if old := rv.MapIndex(k); old.IsValid() {
				v.Set(old)
			}
			err = d.decode(v)
			if err != nil {
				return err
			}

			if checkKey || extKey {
				if err = decodingutil.CheckMapKeyValue(k); err != nil {
					return err
				}
			}
			rv.SetMapIndex(k, v)
		}
		d.depth--

	case reflect.Struct:
		err := d.setStruct(code, rv, k)
		if err != nil {
			return err
		}

	case reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(t.Elem()))
		}

		err := d.decodeWithCode(code, rv.Elem())
		if err != nil {
			return err
		}

	case reflect.Interface:
		elem := rv.Elem()
		switch {
		case !elem.IsValid():
			if t.NumMethod() != 0 {
				return fmt.Errorf("%w: cannot decode into non-empty interface %v", def.ErrUnsupportedType, t)
			}
			v, err := d.asInterfaceWithCode(code, k)
			if err != nil {
				return err
			}
			if v != nil {
				rv.Set(reflect.ValueOf(v))
			}

		case elem.Kind() == reflect.Ptr && !elem.IsNil():
			// decode into the value the interface points to
			return d.decodeWithCode(code, elem.Elem())

		default:
			// decode into a copy of the value the interface holds
			n := reflect.New(elem.Type()).Elem()
			n.Set(elem)
			if err := d.decodeWithCode(code, n); err != nil {
				return err
			}
			rv.Set(n)
		}

	default:
		return fmt.Errorf("%v is %w type", rv.Kind(), def.ErrUnsupportedType)
	}
	return nil
}

// enter increments the nesting depth. Callers decrement d.depth on success;
// on error the whole decode is aborted, so no decrement is needed.
func (d *decoder) enter() error {
	d.depth++
	if d.depth > decodingutil.MaxDepth {
		return decodingutil.DepthError()
	}
	return nil
}

func (d *decoder) errorTemplate(code byte, k reflect.Kind) error {
	return fmt.Errorf("%w %x decoding as %v", def.ErrCanNotDecode, code, k)
}
