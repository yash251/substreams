package goja

import (
	"unicode/utf8"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
	"golang.org/x/text/encoding/unicode"
)

// Not really correlated to text encoding, more to Goja, but shared here for convenience
var (
	symApi = goja.NewSymbol("api")
)

func getEncoding(vm *goja.Runtime, enc goja.Value) (codec TextEncoding) {
	if !goja.IsUndefined(enc) {
		codec = textEncodings[enc.String()]
		if codec == nil {
			panic(errors.NewTypeError(vm, "ERR_UNKNOWN_ENCODING", "Unknown encoding: %s", enc))
		}
	} else {
		codec = utf8TextEncoding
	}
	return
}

var textEncodings = map[string]TextEncoding{
	"utf8":  utf8TextEncoding,
	"utf-8": utf8TextEncoding,
}

type TextEncoding interface {
	Name() string
	Decode(*goja.Runtime, []byte) string
	Encode(*goja.Runtime, []byte) string
}

var utf8TextEncoding TextEncoding = _utf8TextEncoding{}

type _utf8TextEncoding struct{}

func (_utf8TextEncoding) Name() string {
	return "utf-8"
}

func (_utf8TextEncoding) Decode(vm *goja.Runtime, b []byte) string {
	if !utf8.Valid(b) {
		panic(errors.NewTypeError(vm, "ERR_INVALID_DATA", "The \"input\" argument contains invalid UTF-8 data point(s)."))
	}

	return string(b)
}

func (_utf8TextEncoding) Encode(vm *goja.Runtime, b []byte) string {
	r, _ := unicode.UTF8.NewDecoder().Bytes(b)
	return string(r)
}

func expandSlice(b []byte, l int) (dst, res []byte) {
	if cap(b)-len(b) < l {
		b1 := make([]byte, len(b)+l)
		copy(b1, b)
		dst = b1[len(b):]
		res = b1
	} else {
		dst = b[len(b) : len(b)+l]
		res = b[:len(b)+l]
	}
	return
}
