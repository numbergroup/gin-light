// Package msgpack is a MessagePack encoder and decoder derived from
// github.com/shamaton/msgpack (MIT, see LICENSE). It has been changed to
// produce and accept the same data as github.com/ugorji/go/codec's
// MsgpackHandle with its default options, which gin used previously:
//
//   - struct field names come from the "codec" tag, then the "json" tag, then
//     the Go field name, and anonymous struct fields are flattened;
//   - strings, byte slices and time.Time use the raw (str) format, without str8;
//   - encoding.BinaryMarshaler and encoding.BinaryUnmarshaler are honoured;
//   - values decoded into an empty interface use int64, uint64, float64,
//     []byte, []any and map[any]any.
package msgpack

import (
	"io"

	"github.com/numbergroup/gin-light/internal/msgpack/decoding"
	"github.com/numbergroup/gin-light/internal/msgpack/encoding"
)

// Encode writes the MessagePack encoding of v to w.
func Encode(w io.Writer, v any) error {
	return encoding.Encode(w, v)
}

// Decode reads a MessagePack value from r and stores it in the value pointed
// to by v.
func Decode(r io.Reader, v any) error {
	return decoding.Decode(r, v)
}
