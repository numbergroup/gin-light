package common

import (
	"reflect"
	"strings"
	"sync"
)

// Field describes a struct field that takes part in encoding/decoding.
type Field struct {
	Name      string
	Index     []int
	OmitEmpty bool
}

var fieldCache sync.Map // map[reflect.Type][]Field

// Fields returns the encodable fields of the struct type t, in declaration
// order. It follows the rules of ugorji/go/codec: the name comes from the
// "codec" tag, then the "json" tag, then the Go field name; anonymous struct
// fields without a tag name are flattened into the parent; and when names
// collide the shallowest field wins, then the first declared.
func Fields(t reflect.Type) []Field {
	if f, ok := fieldCache.Load(t); ok {
		return f.([]Field)
	}
	var all []Field
	var depths []int
	collectFields(t, nil, 0, map[reflect.Type]bool{}, &all, &depths)

	seen := make(map[string]int, len(all))
	keep := make([]bool, len(all))
	for i, f := range all {
		j, dup := seen[f.Name]
		if !dup {
			seen[f.Name] = i
			keep[i] = true
			continue
		}
		if depths[i] < depths[j] {
			keep[j] = false
			seen[f.Name] = i
			keep[i] = true
		}
	}
	fields := make([]Field, 0, len(all))
	for i, f := range all {
		if keep[i] {
			fields = append(fields, f)
		}
	}
	f, _ := fieldCache.LoadOrStore(t, fields)
	return f.([]Field)
}

func collectFields(t reflect.Type, index []int, depth int, visiting map[reflect.Type]bool, all *[]Field, depths *[]int) {
	if visiting[t] {
		return
	}
	visiting[t] = true
	defer delete(visiting, t)

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag, ok := sf.Tag.Lookup("codec")
		if !ok {
			tag = sf.Tag.Get("json")
		}
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")

		idx := make([]int, len(index)+1)
		copy(idx, index)
		idx[len(index)] = i

		if sf.Anonymous && name == "" {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				collectFields(ft, idx, depth+1, visiting, all, depths)
				continue
			}
		}
		if !sf.IsExported() {
			continue
		}
		if name == "" {
			name = sf.Name
		}
		*all = append(*all, Field{
			Name:      name,
			Index:     idx,
			OmitEmpty: hasOption(opts, "omitempty"),
		})
		*depths = append(*depths, depth)
	}
}

func hasOption(opts, want string) bool {
	for opts != "" {
		var o string
		o, opts, _ = strings.Cut(opts, ",")
		if o == want {
			return true
		}
	}
	return false
}
