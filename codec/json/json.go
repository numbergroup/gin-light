// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !go_json

package json

import (
	"encoding/json"
	"io"
)

// Package indicates what library is being used for JSON encoding.
const Package = "encoding/json"

func init() {
	API = jsonApi{}
}

type jsonApi struct{}

func (j jsonApi) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonApi) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (j jsonApi) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func (j jsonApi) NewEncoder(writer io.Writer) Encoder {
	return json.NewEncoder(writer)
}

func (j jsonApi) NewDecoder(reader io.Reader) Decoder {
	return json.NewDecoder(reader)
}
