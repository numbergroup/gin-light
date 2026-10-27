// Copyright 2026 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package msgpack

import (
	"bytes"
	"encoding/hex"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The expected bytes and values below were produced by
// github.com/ugorji/go/codec v1.3.1 with a default MsgpackHandle, which gin
// used before this package.

type binaryValue struct{ X int }

func (b binaryValue) MarshalBinary() ([]byte, error) { return []byte("bin"), nil }
func (b *binaryValue) UnmarshalBinary(p []byte) error {
	b.X = len(p)
	return nil
}

type binaryPointer struct{ X int }

func (b *binaryPointer) MarshalBinary() ([]byte, error) { return []byte("pbin"), nil }
func (b *binaryPointer) UnmarshalBinary(p []byte) error {
	b.X = len(p) + 100
	return nil
}

type textValue struct{ X int }

func (t textValue) MarshalText() ([]byte, error) { return []byte("txt"), nil }

type (
	myByte byte
	MyInt  int
)

type inner struct {
	A int
	B string
}

type (
	e1       struct{ X, Y int }
	e2       struct{ X, Z int }
	deep     struct{ Y int }
	deepWrap struct{ deep }
	lower    struct{ L int }
)

type conflict struct {
	e1
	e2
	deepWrap
}

// Embedded is exported so that a nil *Embedded can be allocated.
type Embedded struct {
	A int
	B string
}

type ptrEmbed struct {
	*Embedded
	C int
}

type hasLower struct {
	lower
	N int
}

type tagWins struct {
	e1
	X2 int `json:"X"`
}

type tagged struct {
	JSON  string `json:"j"`
	Codec string `codec:"c" json:"cj"`
	MP    string `msgpack:"m"`
	Omit  string `json:"o,omitempty"`
	Skip  string `json:"-"`
	Dash  string `json:"-,"` //nolint:staticcheck // a field named "-"
	Empty string `json:",omitempty"`
	unexp int
	MyInt
}

type omit struct {
	S  []int          `json:"s,omitempty"`
	M  map[string]int `json:"m,omitempty"`
	P  *int           `json:"p,omitempty"`
	PZ *int           `json:"pz,omitempty"`
	St inner          `json:"st,omitempty"`
	T  time.Time      `json:"t,omitempty"`
	I  any            `json:"i,omitempty"`
	F  float64        `json:"f,omitempty"`
	B  bool           `json:"b,omitempty"`
	A  [2]int         `json:"a,omitempty"`
	Bs []byte         `json:"bs,omitempty"`
}

type named struct {
	Name string `json:"name"`
}

func intPtr(i int) *int { return &i }

func encodeHex(t *testing.T, v any) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, v))
	return hex.EncodeToString(buf.Bytes())
}

func TestEncode(t *testing.T) {
	tm := time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC)
	tests := []struct {
		in   any
		want string
	}{
		{nil, "c0"},
		{0, "00"},
		{127, "7f"},
		{128, "d10080"},
		{255, "d100ff"},
		{256, "d10100"},
		{65536, "d200010000"},
		{int64(1) << 32, "d30000000100000000"},
		{-1, "ff"},
		{-32, "e0"},
		{-33, "d0df"},
		{-129, "d1ff7f"},
		{-40000, "d2ffff63c0"},
		{int64(math.MinInt64), "d38000000000000000"},
		{uint8(5), "05"},
		{uint(200), "ccc8"},
		{uint64(math.MaxUint64), "cfffffffffffffffff"},
		{int8(-5), "fb"},
		{float32(1.5), "ca3fc00000"},
		{1.5, "cb3ff8000000000000"},
		{true, "c3"},
		{false, "c2"},
		{"", "a0"},
		{"a", "a161"},
		{strings.Repeat("a", 31), "bf" + strings.Repeat("61", 31)},
		{strings.Repeat("a", 32), "da0020" + strings.Repeat("61", 32)},
		{[]byte{1, 2}, "a20102"},
		{[2]byte{1, 2}, "a20102"},
		{[]myByte{1, 2}, "920102"},
		{[]int{1, 2}, "920102"},
		{[]string{"a"}, "91a161"},
		{[]any{1, "a", nil}, "9301a161c0"},
		{map[string]int{"a": 1}, "81a16101"},
		{map[int]string{1: "a"}, "8101a161"},
		{[]int(nil), "c0"},
		{map[string]int(nil), "c0"},
		{(*int)(nil), "c0"},
		{intPtr(5), "05"},
		{complex(1, 0), "cb3ff0000000000000"},
		{tm, "a46553f100"},
		{time.Unix(1700000000, 5).UTC(), "a8000000146553f100"},
		{time.Unix(1<<35, 5).UTC(), "ac000000050000000800000000"},
		{time.Unix(-5, 0).UTC(), "ac00000000fffffffffffffffb"},
		{tm.In(time.FixedZone("X", 3600)), "a46553f100"},
		{time.Time{}, "c0"},
		{time.Second, "d23b9aca00"},
		{binaryValue{1}, "a362696e"},
		{&binaryValue{1}, "a362696e"},
		{binaryPointer{1}, "a47062696e"},
		{textValue{1}, "81a15801"},
		{inner{1, "x"}, "82a14101a142a178"},
		{[]inner{{1, "a"}}, "9182a14101a142a161"},
		{conflict{e1{1, 2}, e2{3, 4}, deepWrap{deep{5}}}, "83a15801a15902a15a04"},
		{ptrEmbed{C: 1}, "83a141c0a142c0a14301"},
		{ptrEmbed{&Embedded{1, "z"}, 3}, "83a14101a142a17aa14303"},
		{hasLower{lower{1}, 2}, "82a14c01a14e02"},
		{tagWins{e1{1, 2}, 3}, "82a15902a15803"},
		{tagged{JSON: "j", Codec: "c", MP: "m", Skip: "s", Dash: "d", MyInt: 3}, "85a16aa16aa163a163a24d50a16da12da164a54d79496e7403"},
		{tagged{Omit: "o", Empty: "e"}, "87a16aa0a163a0a24d50a0a16fa16fa12da0a5456d707479a165a54d79496e7400"},
		{omit{}, "80"},
		{omit{S: []int{}, M: map[string]int{}, PZ: intPtr(0), Bs: []byte{}}, "84a17390a16d80a2707a00a26273a0"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, encodeHex(t, tt.in), "%#v", tt.in)
	}
}

func TestEncodeErrors(t *testing.T) {
	for _, v := range []any{complex(1, 2), make(chan int)} {
		require.Error(t, Encode(&bytes.Buffer{}, v), "%#v", v)
	}
}

func TestDecode(t *testing.T) {
	tm := time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC)
	tests := []struct {
		in   string
		into any
		want any
	}{
		{"7f", new(any), int64(127)},
		{"d10080", new(any), int64(128)},
		{"cc05", new(any), uint64(5)},
		{"ca3fc00000", new(any), 1.5},
		{"a161", new(any), []byte("a")},
		{"c40161", new(any), []byte("a")},
		{"920102", new(any), []any{int64(1), int64(2)}},
		{"81a16101", new(any), map[any]any{"a": int64(1)}},
		{"82a16101a162a163", new(map[string]any), map[string]any{"a": int64(1), "b": []byte("c")}},
		{"d6ff6553f100", new(any), tm},
		{"a46553f100", new(time.Time), tm},
		{"a8000000146553f100", new(time.Time), tm.Add(5)},
		{"d6ff6553f100", new(time.Time), tm},
		{"c70cff000000050000000000000005", new(time.Time), time.Unix(5, 5).UTC()},
		{"c0", new(time.Time), time.Time{}},
		{"c403616263", new(string), "abc"},
		{"a3616263", new([]byte), []byte("abc")},
		{"920102", new(string), "\x01\x02"},
		{"920102", new([]byte), []byte{1, 2}},
		{"81a16101", new([]any), []any{[]byte("a"), int64(1)}},
		{"810101", new([3]int), [3]int{1, 1, 0}},
		{"93010203", new([1]int), [1]int{1}},
		{"cb4008000000000000", new(int), 3},
		{"cb4008000000000000", new(uint8), uint8(3)},
		{"cfffffffffffffffff", new(int), -1},
		{"cfffffffffffffffff", new(float64), float64(-1)},
		{"01", new(bool), true},
		{"05", new(time.Duration), time.Duration(5)},
		{"81a46e616d65a178", new(named), named{Name: "x"}},
		{"81a44e616d65a178", new(named), named{}},
		{"81a2453107", new(conflict), conflict{}},
		{"81a15807", new(conflict), conflict{e1: e1{X: 7}}},
		{"81a14101", new(ptrEmbed), ptrEmbed{Embedded: &Embedded{A: 1}}},
		{"81a141c0", new(ptrEmbed), ptrEmbed{}},
		{"81a14c07", new(hasLower), hasLower{lower: lower{L: 7}}},
		{"9201a178", new(inner), inner{A: 1, B: "x"}},
		{"81c0a161", new(inner), inner{}},
		{"81a3787878a178", new(inner), inner{}},
		{"a3616263", new(binaryValue), binaryValue{X: 3}},
		{"a3616263", new(binaryPointer), binaryPointer{X: 103}},
		{"c0", new(inner), inner{}},
	}
	for _, tt := range tests {
		data, err := hex.DecodeString(tt.in)
		require.NoError(t, err)
		require.NoError(t, Decode(bytes.NewReader(data), tt.into), "%s into %T", tt.in, tt.into)
		assert.Equal(t, tt.want, reflect.ValueOf(tt.into).Elem().Interface(), "%s into %T", tt.in, tt.into)
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		in   string
		into any
	}{
		{"cb3ff8000000000000", new(int)},
		{"cd012c", new(uint8)},
		{"ff", new(uint8)},
		{"ccc8", new(int8)},
		{"cb7e37e43c8800759c", new(float32)},
		{"02", new(bool)},
		{"01", new(string)},
		{"a56162636465", new([3]byte)},
		{"810101", new([]byte)},
		{"81a16101", new(map[int]string)},
		{"a56162636465", new(time.Time)},
		{"a161", new([]any)},
		{"", new(any)},
		{"82a14101", new(inner)},
		{"81a141c1", new(inner)},
	}
	for _, tt := range tests {
		data, err := hex.DecodeString(tt.in)
		require.NoError(t, err)
		require.Error(t, Decode(bytes.NewReader(data), tt.into), "%s into %T", tt.in, tt.into)
	}
}

func TestDecodeIntoExisting(t *testing.T) {
	decode := func(in string, v any) {
		t.Helper()
		data, err := hex.DecodeString(in)
		require.NoError(t, err)
		require.NoError(t, Decode(bytes.NewReader(data), v))
	}

	m := map[string]int{"old": 1}
	decode("81a16101", &m)
	assert.Equal(t, map[string]int{"old": 1, "a": 1}, m)

	a := [3]int{9, 9, 9}
	decode("9101", &a)
	assert.Equal(t, [3]int{1, 9, 9}, a)

	s := inner{A: 1, B: "keep"}
	decode("81a14105", &s)
	assert.Equal(t, inner{A: 5, B: "keep"}, s)

	var held any = inner{9, "x"}
	decode("81a14101", &held)
	assert.Equal(t, inner{1, "x"}, held)

	in := inner{9, "x"}
	var ptr any = &in
	decode("81a14101", &ptr)
	assert.Equal(t, inner{1, "x"}, in)

	decode("c0", &ptr)
	assert.Nil(t, ptr)

	p := intPtr(1)
	decode("c0", &p)
	assert.Nil(t, p)

	dup := map[string]inner{}
	decode("82a17881a14101a17881a142a162", &dup)
	assert.Equal(t, map[string]inner{"x": {1, "b"}}, dup)
}

func TestRoundTrip(t *testing.T) {
	type all struct {
		I   int
		U   uint64
		F   float32
		S   string
		Bs  []byte
		T   time.Time
		M   map[string][]int
		P   *inner
		In  inner
		Any any
		E   ptrEmbed
	}
	v := all{
		I: -1 << 40, U: math.MaxUint64, F: 2.5, S: strings.Repeat("s", 300), Bs: bytes.Repeat([]byte{7}, 70000),
		T: time.Unix(1<<35, 123).UTC(), M: map[string][]int{"a": {1, 2}}, P: &inner{1, "p"}, In: inner{2, "i"},
		Any: map[any]any{"k": int64(-3)}, E: ptrEmbed{Embedded: &Embedded{3, "e"}, C: 4},
	}
	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, v))
	var got all
	require.NoError(t, Decode(&buf, &got))
	assert.Equal(t, v, got)
}
