package decoding

import (
	"reflect"

	"github.com/gin-gonic/gin/internal/msgpack/def"
)

func (d *decoder) asBoolWithCode(code byte, k reflect.Kind) (bool, error) {
	switch code {
	case def.True:
		return true, nil
	case def.False, 0x00:
		return false, nil
	case 0x01:
		// ugorji/go/codec also accepts the fixints 0 and 1
		return true, nil
	}
	return false, d.errorTemplate(code, k)
}
