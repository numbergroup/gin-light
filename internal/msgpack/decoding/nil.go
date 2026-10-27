package decoding

import "github.com/gin-gonic/gin/internal/msgpack/def"

func (d *decoder) isCodeNil(v byte) bool {
	return def.Nil == v
}
