package encoding

import (
	"reflect"

	"github.com/gin-gonic/gin/internal/msgpack/common"
)

func (e *encoder) writeStruct(rv reflect.Value) error {
	fields := common.Fields(rv.Type())

	values := make([]reflect.Value, len(fields))
	l := 0
	for i := range fields {
		v, ok := fieldByIndex(rv, fields[i].Index)
		values[i] = v
		if ok && fields[i].OmitEmpty && isEmptyValue(v) {
			continue
		}
		l++
	}

	if err := e.writeMapLength(l); err != nil {
		return err
	}

	for i := range fields {
		v := values[i]
		if v.IsValid() && fields[i].OmitEmpty && isEmptyValue(v) {
			continue
		}
		if err := e.writeString(fields[i].Name); err != nil {
			return err
		}
		if err := e.create(v); err != nil {
			return err
		}
	}
	return nil
}

// fieldByIndex is like reflect.Value.FieldByIndex but returns an invalid
// value (encoded as nil) when it passes through a nil embedded pointer.
func fieldByIndex(rv reflect.Value, index []int) (reflect.Value, bool) {
	for i, x := range index {
		if i > 0 && rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return reflect.Value{}, false
			}
			rv = rv.Elem()
		}
		rv = rv.Field(x)
	}
	return rv, true
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Struct:
		return v.IsZero()
	case reflect.String:
		return v.Len() == 0
	case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return v.IsZero()
}
