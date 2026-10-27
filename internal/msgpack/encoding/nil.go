package encoding

import "github.com/gin-gonic/gin/internal/msgpack/def"

func (e *encoder) writeNil() error {
	return e.setByte1Int(def.Nil)
}
