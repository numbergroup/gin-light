package encoding

import "github.com/numbergroup/gin-light/internal/msgpack/def"

func (e *encoder) writeNil() error {
	return e.setByte1Int(def.Nil)
}
