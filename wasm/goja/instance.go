package goja

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/dop251/goja_nodejs/util"
	"github.com/streamingfast/substreams/wasm"
	"go.uber.org/zap"
)

var _ wasm.Instance = (*instance)(nil)

type instance struct {
	CurrentCall *wasm.Call

	vm      *goja.Runtime
	exports *goja.Object

	vmTypeUint8Array goja.Value

	isClosed bool
}

func newInstance(ctx context.Context, program *goja.Program) (*instance, error) {
	if tracer.Enabled() {
		zlog.Debug("instantiating new instance")
		defer zlog.Debug("instantiated new instance")
	}

	// Create a new VM which is essentially a new global context that the whole JavaScript
	// code executed will share. The order in which we register the builtins and other functions
	// is important, re-ordering the code pieces must be done with great care and testing.
	vm := goja.New()

	// Assigns "module" to represent the global context, `module.exports` is used by
	// the bundle to export functions so we must provide `module` early on.
	if err := vm.Set("module", vm.GlobalObject()); err != nil {
		return nil, fmt.Errorf("setting exports: %w", err)
	}

	inst := &instance{
		vm:               vm,
		vmTypeUint8Array: vm.Get("Uint8Array"),
	}

	if err := inst.registerBuiltIns(); err != nil {
		return nil, fmt.Errorf("registering builtins: %w", err)
	}

	if err := inst.registerIntrinsics(); err != nil {
		return nil, fmt.Errorf("registering intrinsics: %w", err)
	}

	// Load the script into our VM so that user functions and other data are loaded
	_, err := vm.RunProgram(program)
	if err != nil {
		return nil, fmt.Errorf("evaluating javascript code: %w", err)
	}

	// The bundle uses `module.exports = ...` to export the functions, thus, we must
	// extract the `exports` object from the global context (`module` is a symlink to the global context).
	//
	// This will be the list of functions exported by the script on which we can rely to find entry points.
	inst.exports = vm.Get("exports").ToObject(vm)

	return inst, nil
}

// Cleanup implements wasm.Instance.
func (*instance) Cleanup(ctx context.Context) error {
	return nil
}

func (i *instance) Close(ctx context.Context) error {
	i.isClosed = true
	return nil
}

// func (i *instance) newExtensionFunction(ctx context.Context, namespace, name string, f wasm.WASMExtension) interface{} {
// 	return func(ptr, length, outputPtr int32) {
// 		data := i.Heap.ReadBytes(ptr, length)

// 		out, err := f(ctx, reqctx.Details(ctx).UniqueIDString(), i.CurrentCall.Clock, data)
// 		if err != nil {
// 			panic(fmt.Errorf(`running wasm extension "%s::%s": %w`, namespace, name, err))
// 		}

// 		// It's unclear if WASMExtension implementor will correctly handle the context canceled case, as a safety
// 		// measure, we check if the context was canceled without being handled correctly and stop here.
// 		if ctx.Err() == context.Canceled {
// 			panic(fmt.Errorf("running wasm %s@%s extension has been stop upstream in the call stack: %w", namespace, name, ctx.Err()))
// 		}

// 		if err = writeOutputToHeap(i, outputPtr, out); err != nil {
// 			panic(fmt.Errorf("write output to heap %w", err))
// 		}
// 	}
// }

func (i *instance) registerBuiltIns() error {
	bufferModule := i.newJSModule()
	buffer.Require(i.vm, bufferModule)

	utilModule := i.newJSModule()
	util.Require(i.vm, utilModule)

	textModule := i.newJSModule()
	RequireTextEncoder(i.vm, textModule)
	RequireTextDecoder(i.vm, textModule)

	// Must come after `util`!
	consoleModule := i.newJSModule()
	requireConsole(i.vm, utilModule.Get("exports").ToObject(i.vm), consoleModule)

	i.exposeModuleExportsGlobally("node:buffer", bufferModule)
	i.exposeModuleExportsGlobally("node:util", utilModule)
	i.exposeModuleExportsGlobally("node:text", textModule)
	i.exposeModuleGlobally("console", consoleModule)

	return nil
}

func (i *instance) newJSModule() *goja.Object {
	module := i.vm.NewObject()
	module.Set("exports", i.vm.NewObject())

	return module
}

// exposeModuleGlobally exposes the module as a global variable, the `moduleName` variable
// is exposed and represents the `module.exports` object.
//
// Taking `console` module as an example, on `exports` we have defined `log`, `warn`, and `error.
// After using this method, `console.log`, `console.warn` and `console.error` will be available
// from JavaScript code.
func (i *instance) exposeModuleGlobally(moduleName string, module *goja.Object) error {
	if tracer.Enabled() {
		zlog.Debug("exposing module globally", zap.String("module", moduleName))
	}

	if err := i.vm.Set(moduleName, module.Get("exports")); err != nil {
		return fmt.Errorf("setting module '%s' on global scope: %w", moduleName, err)
	}

	return nil
}

// exposeModuleExportsGlobally exposes the exports of the module as a global variable,
// the `module.exports` is .
//
// Taking `console` module as an example, on `exports` we have defined `log`, `warn`, and `error.
// After using this method, `console.log`, `console.warn` and `console.error` will be available
// from JavaScript code.
func (i *instance) exposeModuleExportsGlobally(moduleName string, module *goja.Object) error {
	if tracer.Enabled() {
		zlog.Debug("exposing module exports globally", zap.String("module", moduleName))
	}

	exports := module.Get("exports").ToObject(i.vm)
	for _, export := range exports.Keys() {
		if tracer.Enabled() {
			zlog.Debug("registering builtin", zap.String("module", moduleName), zap.String("export", export))
		}

		if err := i.vm.Set(export, exports.Get(export)); err != nil {
			return fmt.Errorf("setting module export '%s' on global scope: %w", export, err)
		}
	}

	return nil
}

func (i *instance) registerIntrinsics() error {
	object := i.vm.NewObject()

	object.Set("output", func(v goja.Value) {
		data, err := i.toGolangBytes(v)
		if err != nil {
			i.vm.Interrupt(err)
			return
		}

		i.CurrentCall.SetReturnValue(data)
	})

	if err := i.vm.Set("substreams_engine", object); err != nil {
		return fmt.Errorf("setting substreams_engine: %w", err)
	}

	return nil
}

func (i *instance) toGolangBytes(from goja.Value) ([]byte, error) {
	obj := from.ToObject(i.vm)
	switch obj.ClassName() {
	case "String":
		in := obj.String()
		in = strings.TrimPrefix(in, "0x")
		in = strings.TrimPrefix(in, "0X")

		if len(in)%2 != 0 {
			in = "0" + in
		}

		return hex.DecodeString(in)

	case "Array":
		var b []byte
		if err := i.vm.ExportTo(from, &b); err != nil {
			return nil, err
		}
		return b, nil

	case "Object":
		if !obj.Get("constructor").SameAs(i.vmTypeUint8Array) {
			break
		}

		b := obj.Export().([]byte)
		return b, nil
	}
	return nil, errors.New("invalid buffer type")
}

func (i *instance) toJSUint8Array(data []byte) (goja.Value, error) {
	out, err := i.vm.New(i.vmTypeUint8Array, i.vm.ToValue(i.vm.NewArrayBuffer(data)))
	if err != nil {
		return nil, fmt.Errorf("creating JavaScript Uint8Array: %w", err)
	}

	return out, nil
}

// 	err := i.registerLoggerImports(linker)
// 	if err != nil {
// 		return fmt.Errorf("registering logger imports: %w", err)
// 	}
// 	err = i.registerStateImports(linker)
// 	if err != nil {
// 		return fmt.Errorf("registering state imports: %w", err)
// 	}

// 	if err = linker.FuncWrap("env", "register_panic",
// 		func(msgPtr, msgLength int32, filenamePtr, filenameLength int32, lineNumber, columnNumber int32, caller *wasmtime.Caller) {
// 			message := i.Heap.ReadString(msgPtr, msgLength)

// 			var filename string
// 			if filenamePtr != 0 {
// 				filename = i.Heap.ReadString(filenamePtr, filenameLength)
// 			}

// 			i.CurrentCall.SetPanicError(message, filename, int(lineNumber), int(columnNumber))
// 		},
// 	); err != nil {
// 		return fmt.Errorf("registering panic import: %w", err)
// 	}

// 	return nil
// }

// func (i *instance) registerLoggerImports(linker *wasmtime.Linker) error {
// 	if err := linker.FuncWrap("logger", "println",
// 		func(ptr int32, length int32) {
// 			message := i.Heap.ReadString(ptr, length)
// 			i.CurrentCall.AppendLog(message)
// 		},
// 	); err != nil {
// 		return fmt.Errorf("registering println import: %w", err)
// 	}
// 	return nil
// }

// func (i *instance) registerStateImports(linker *wasmtime.Linker) error {
// 	functions := map[string]interface{}{}
// 	functions["set"] = i.set
// 	functions["set_if_not_exists"] = i.setIfNotExists
// 	functions["append"] = i.append
// 	functions["delete_prefix"] = i.deletePrefix
// 	functions["add_bigint"] = i.addBigInt
// 	functions["add_bigdecimal"] = i.addBigDecimal
// 	functions["add_bigfloat"] = i.addBigDecimal
// 	functions["add_int64"] = i.addInt64
// 	functions["add_float64"] = i.addFloat64
// 	functions["set_min_int64"] = i.setMinInt64
// 	functions["set_min_bigint"] = i.setMinBigint
// 	functions["set_min_float64"] = i.setMinFloat64
// 	functions["set_min_bigdecimal"] = i.setMinBigDecimal
// 	functions["set_min_bigfloat"] = i.setMinBigDecimal
// 	functions["set_max_int64"] = i.setMaxInt64
// 	functions["set_max_bigint"] = i.setMaxBigInt
// 	functions["set_max_float64"] = i.setMaxFloat64
// 	functions["set_max_bigdecimal"] = i.setMaxBigDecimal
// 	functions["set_max_bigfloat"] = i.setMaxBigDecimal
// 	functions["get_at"] = i.getAt
// 	functions["get_first"] = i.getFirst
// 	functions["get_last"] = i.getLast
// 	functions["has_at"] = i.hasAt
// 	functions["has_first"] = i.hasFirst
// 	functions["has_last"] = i.hasLast

// 	for n, f := range functions {
// 		if err := linker.FuncWrap("state", n, f); err != nil {
// 			return fmt.Errorf("registering %s import: %w", n, err)
// 		}
// 	}

// 	return nil
// }
