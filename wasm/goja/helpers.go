package goja

import (
	"reflect"

	"github.com/dop251/goja"
)

var (
	reflectTypeString = reflect.TypeOf("")
)

func isString(v goja.Value) bool {
	return v.ExportType() == reflectTypeString
}

func bytesFromValue(r *goja.Runtime, v goja.Value) []byte {
	var b []byte
	err := r.ExportTo(v, &b)
	if err != nil {
		return []byte(v.String())
	}
	return b
}
