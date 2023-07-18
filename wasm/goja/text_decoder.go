package goja

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
)

func RequireTextDecoder(runtime *goja.Runtime, module *goja.Object) {
	b := &TextDecoder{r: runtime}
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
	b.textDecoderCtorObj = ctor
	b.uint8ArrayCtorObj = uint8ArrayObj

	proto := runtime.NewObject()
	proto.SetPrototype(uint8ArrayObj.Get("prototype").ToObject(runtime))
	proto.DefineDataProperty("constructor", ctor, goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
	proto.Set("decode", b.proto_decode)

	ctor.Set("prototype", proto)
	// ctor.Set("poolSize", 8192)
	// ctor.Set("from", b.from)
	// ctor.Set("alloc", b.alloc)

	exports := module.Get("exports").(*goja.Object)
	exports.Set("TextDecoder", ctor)
}

type TextDecoder struct {
	r *goja.Runtime

	textDecoderCtorObj *goja.Object

	uint8ArrayCtorObj *goja.Object
	uint8ArrayCtor    goja.Constructor
}

func (b *TextDecoder) ctor(call goja.ConstructorCall) (res *goja.Object) {
	encoding := call.Argument(0)

	textEncoding := utf8TextEncoding
	if !goja.IsUndefined(encoding) {
		if !isString(encoding) {
			panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "encoding", "string", encoding))
		}

		var found bool
		textEncoding, found = textEncodings[encoding.ToString().String()]
		if !found {
			panic(errors.NewTypeError(b.r, "ERR_NOT_IMPLEMENTED", "encoding "+encoding.ToString().String()+" is not implemented"))
		}
	}

	call.This.Set("encoding", textEncoding.Name())
	call.This.Set("fatal", false)
	call.This.Set("ignoreBOM", false)

	return call.This
}

// Returns the result of running encoding’s decoder. The method can be invoked zero or more
// times with options’s stream set to true, and then once without options’s stream (or set to false),
// to process a fragmented input. If the invocation without options’s stream (or set to false) has no
// input, it’s clearest to omit both arguments.
//
//	var string = "", decoder = new TextDecoder(encoding), buffer;
//	while(buffer = next_chunk()) {
//	  string += decoder.decode(buffer, {stream:true});
//	}
//	string += decoder.decode(); // end-of-queue
//
// If the error mode is "fatal" and encoding’s decoder returns error, throws a TypeError.
//
// See https://encoding.spec.whatwg.org/#interface-textdecoder
func (b *TextDecoder) proto_decode(call goja.FunctionCall) goja.Value {
	other := call.Argument(0)

	switch {
	case b.r.InstanceOf(other, b.uint8ArrayCtorObj):
		inputBytes := bytesFromValue(b.r, other)
		encoding := getEncoding(b.r, call.This.ToObject(b.r).Get("encoding").ToString())

		return b.r.ToValue(encoding.Decode(b.r, inputBytes))

		// FIXME: Does `decode` support `Buffer` and other Array-Like input?
	}

	panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"input\" argument must be an instance of Uint8Array."))
}
