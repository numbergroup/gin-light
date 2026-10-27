package encoding

import (
	"reflect"
)

var typeByte = reflect.TypeOf(byte(0))

func (e *encoder) isByteSlice(rv reflect.Value) bool {
	return rv.Type().Elem() == typeByte
}
