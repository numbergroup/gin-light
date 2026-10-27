package encoding

import "github.com/numbergroup/gin-light/internal/msgpack/def"

func (e *encoder) writeBool(v bool) error {
	if v {
		return e.setByte1Int(def.True)
	}
	return e.setByte1Int(def.False)
}
