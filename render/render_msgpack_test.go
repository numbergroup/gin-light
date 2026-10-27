// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package render

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin/internal/msgpack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMsgPack(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	(MsgPack{data}).WriteContentType(w)
	assert.Equal(t, "application/msgpack; charset=utf-8", w.Header().Get("Content-Type"))

	err := (MsgPack{data}).Render(w)

	require.NoError(t, err)

	var decoded map[string]string
	err = msgpack.Decode(bytes.NewReader(w.Body.Bytes()), &decoded)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"foo": "bar"}, decoded)
	assert.Equal(t, "application/msgpack; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestWriteMsgPack(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
		"num": 42,
	}

	err := WriteMsgPack(w, data)
	require.NoError(t, err)

	assert.Equal(t, "application/msgpack; charset=utf-8", w.Header().Get("Content-Type"))

	var decoded struct {
		Foo string `json:"foo"`
		Num int    `json:"num"`
	}
	err = msgpack.Decode(bytes.NewReader(w.Body.Bytes()), &decoded)
	require.NoError(t, err)
	assert.Equal(t, "bar", decoded.Foo)
	assert.Equal(t, 42, decoded.Num)
}

type failWriter struct {
	*httptest.ResponseRecorder
}

func (w *failWriter) Write(data []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestRenderMsgPackError(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	err := (MsgPack{data}).Render(&failWriter{w})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}
