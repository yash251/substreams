package goja

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
)

func RequireTextEncoder(runtime *goja.Runtime, module *goja.Object) {
	b := &TextEncoder{r: runtime}
	uint8Array := runtime.Get("Uint8Array")
	if c, ok := goja.AssertConstructor(uint8Array); ok {
		b.uint8ArrayCtor = c
	} else {
		panic(runtime.NewTypeError("Uint8Array is not a constructor"))
	}
	uint8ArrayObj := uint8Array.ToObject(runtime)

	ctor := runtime.ToValue(b.ctor).ToObject(runtime)
	ctor.SetPrototype(uint8ArrayObj)
	ctor.DefineDataPropertySymbol(symApi, runtime.ToValue(b), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	b.textEncoderCtorObj = ctor
	b.uint8ArrayCtorObj = uint8ArrayObj

	proto := runtime.NewObject()
	proto.SetPrototype(uint8ArrayObj.Get("prototype").ToObject(runtime))
	proto.DefineDataProperty("constructor", ctor, goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
	proto.Set("encode", b.proto_encode)
	proto.Set("encodeInto", b.proto_encodeInto)

	ctor.Set("prototype", proto)

	exports := module.Get("exports").(*goja.Object)
	exports.Set("TextEncoder", ctor)
}

type TextEncoder struct {
	r *goja.Runtime

	textEncoderCtorObj *goja.Object

	uint8ArrayCtorObj *goja.Object
	uint8ArrayCtor    goja.Constructor
}

func (b *TextEncoder) ctor(call goja.ConstructorCall) (res *goja.Object) {
	encoding := call.Argument(0)

	textEncoding := utf8TextEncoding
	if !goja.IsUndefined(encoding) {
		panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "encoder accept no constructor arguments"))
	}

	call.This.Set("encoding", textEncoding.Name())
	call.This.Set("fatal", false)
	call.This.Set("ignoreBOM", false)

	return call.This
}

// Returns the result of running UTF-8’s encoder (e.g. converting a JavaScript string into
// and Uint8Array of UTF-8 valid code points of the said string).
//
// See https://encoding.spec.whatwg.org/#interface-textencoder
func (b *TextEncoder) proto_encode(call goja.FunctionCall) goja.Value {
	input := call.Argument(0)
	if goja.IsUndefined(input) {
		res, err := b.uint8ArrayCtor(b.uint8ArrayCtorObj)
		if err != nil {
			panic(err)
		}

		return res
	}

	switch {
	case isString(input):
		data := []byte(input.ToString().String())

		o, err := b.uint8ArrayCtor(b.uint8ArrayCtorObj, b.r.ToValue(b.r.NewArrayBuffer(data)))
		if err != nil {
			panic(err)
		}

		return o
	}

	panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"input\" argument must be an instance of string."))
}

// Runs the UTF-8 encoder on source, stores the result of that operation into destination, and returns the progress
// made as an object wherein read is the number of converted code units of source and written is the number of bytes
// modified in destination.
//
// See https://encoding.spec.whatwg.org/#textencoder
func (b *TextEncoder) proto_encodeInto(call goja.FunctionCall) goja.Value {
	panic(errors.NewTypeError(b.r, "ERR_NOT_IMPLEMENTED", "The \"encodeInto\" method is not implemented."))
}
