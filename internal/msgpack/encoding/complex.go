package encoding

import (
	"fmt"

	"github.com/gin-gonic/gin/internal/msgpack/def"
)

// Like ugorji/go/codec, complex numbers are encoded as their real part and
// rejected when they have an imaginary part.

func (e *encoder) writeComplex64(v complex64) error {
	if imag(v) != 0 {
		return fmt.Errorf("%w: cannot encode complex number %v with an imaginary part", def.ErrUnsupportedType, v)
	}
	return e.writeFloat32(float64(real(v)))
}

func (e *encoder) writeComplex128(v complex128) error {
	if imag(v) != 0 {
		return fmt.Errorf("%w: cannot encode complex number %v with an imaginary part", def.ErrUnsupportedType, v)
	}
	return e.writeFloat64(real(v))
}
