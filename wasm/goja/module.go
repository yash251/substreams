package goja

import (
	"context"
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/streamingfast/substreams/wasm"
	"go.uber.org/zap"
)

// A Module represents a wazero.Runtime that clears and is destroyed upon completion of a request.
// It has the pre-compiled `env` host module, as well as pre-compiled WASM code provided by the user
type Module struct {
	sync.Mutex

	program *goja.Program
}

func init() {
	wasm.RegisterModuleFactory("goja", wasm.ModuleFactoryFunc(newModule))
}

func newModule(ctx context.Context, javascriptCode []byte, registry *wasm.Registry) (wasm.Module, error) {
	program, err := goja.Compile("substreams", string(javascriptCode), true)
	if err != nil {
		return nil, fmt.Errorf("compiling javascript: %w", err)
	}

	return &Module{
		program: program,
	}, nil
}

func (m *Module) Close(ctx context.Context) error {
	// closeFuncs := []func(context.Context) error{
	// 	m.wazRuntime.Close,
	// 	m.userModule.Close,
	// }
	// for _, hostMod := range m.hostModules {
	// 	closeFuncs = append(closeFuncs, hostMod.Close)
	// }
	// for _, f := range closeFuncs {
	// 	if err := f(ctx); err != nil {
	// 		return err
	// 	}
	// }
	// return nil
	return nil
}

func (m *Module) NewInstance(ctx context.Context) (out wasm.Instance, err error) {
	return newInstance(ctx, m.program)
}

func (m *Module) ExecuteNewCall(ctx context.Context, call *wasm.Call, cachedInstance wasm.Instance, arguments []wasm.Argument) (out wasm.Instance, err error) {
	instGeneric := cachedInstance
	if instGeneric == nil {
		instGeneric, err = newInstance(ctx, m.program)
		if err != nil {
			return nil, fmt.Errorf("new instance: %w", err)
		}
	}

	inst := instGeneric.(*instance)

	if tracer.Enabled() {
		zlog.Debug("goja global exports keys", zap.Strings("keys", inst.exports.Keys()))
	}

	entrypoint := inst.exports.Get(call.Entrypoint)
	if entrypoint == nil {
		return inst, fmt.Errorf(`could not find exported %q function`, call.Entrypoint)
	}

	f, ok := goja.AssertFunction(entrypoint)
	if !ok {
		return inst, fmt.Errorf(`value "exports" was not a function, got %s`, entrypoint.ExportType())
	}

	args := make([]goja.Value, len(arguments))

	// var inputStoreCount int

	for i, input := range arguments {
		switch v := input.(type) {
		case *wasm.StoreWriterOutput:
		case *wasm.StoreReaderInput:
			// FIXME: Add support for store input
			// inputStoreCount++
			// args = append(args, int32(inputStoreCount-1))
		case wasm.ValueArgument:
			if tracer.Enabled() {
				zlog.Debug("turning value argument into Goja Value", zap.String("name", input.Name()), zap.Int("byte_count", len(v.Value())))
			}

			args[i], err = inst.toJSUint8Array(v.Value())
			if err != nil {
				return inst, fmt.Errorf("writing input %q to JavaScript: %w", input.Name(), err)
			}
		default:
			panic("unknown wasm argument type")
		}
	}

	// Set the current call on the instance, so that intrinsic functions can access it
	inst.CurrentCall = call

	_, err = f(goja.Undefined(), args...)
	if err != nil {
		return inst, fmt.Errorf("call: %w", err)
	}

	return inst, nil
}

// func (m *Module) instantiateModule(ctx context.Context) (api.Module, error) {
// 	m.Lock()
// 	defer m.Unlock()

// 	for _, hostMod := range m.hostModules {
// 		if m.wazRuntime.Module(hostMod.Name()) != nil {
// 			continue
// 		}
// 		_, err := m.wazRuntime.InstantiateModule(ctx, hostMod, m.wazModuleConfig.WithName(hostMod.Name()))
// 		if err != nil {
// 			return nil, fmt.Errorf("instantiating host module %q: %w", hostMod.Name(), err)
// 		}
// 	}
// 	mod, err := m.wazRuntime.InstantiateModule(ctx, m.userModule, m.wazModuleConfig.WithName(""))
// 	return mod, err
// }

// func addExtensionFunctions(ctx context.Context, runtime wazero.Runtime, registry *wasm.Registry) (out []wazero.CompiledModule, err error) {
// 	for namespace, imports := range registry.Extensions {
// 		builder := runtime.NewHostModuleBuilder(namespace)
// 		for importName, f := range imports {
// 			builder.NewFunctionBuilder().
// 				WithGoFunction(api.GoFunc(func(ctx context.Context, stack []uint64) {
// 					inst := instanceFromContext(ctx)
// 					ptr, length, outputPtr := uint32(stack[0]), uint32(stack[1]), uint32(stack[2])
// 					data := readBytes(inst, ptr, length)
// 					call := wasm.FromContext(ctx)

// 					t0 := time.Now()
// 					out, err := f(ctx, reqctx.Details(ctx).UniqueIDString(), call.Clock, data)
// 					if err != nil {
// 						panic(fmt.Errorf(`running wasm extension "%s::%s": %w`, namespace, importName, err))
// 					}
// 					reqctx.ReqStats(ctx).RecordWasmExtDuration(importName, time.Since(t0))

// 					if ctx.Err() == context.Canceled {
// 						// Sometimes long-running extensions will come back to a canceled context.
// 						// so avoid writing to memory then
// 						return
// 					}

// 					if err := writeOutputToHeap(ctx, inst, outputPtr, out); err != nil {
// 						panic(fmt.Errorf("write output to heap %w", err))
// 					}
// 				}), []parm{i32, i32, i32}, []parm{}).
// 				Export(importName)
// 		}
// 		mod, err := builder.Compile(ctx)
// 		if err != nil {
// 			return nil, fmt.Errorf("compiling wasm extension %q: %w", namespace, err)
// 		}
// 		out = append(out, mod)
// 	}
// 	return
// }

// func addHostFunctions(ctx context.Context, runtime wazero.Runtime, moduleName string, funcs []funcs) (wazero.CompiledModule, error) {
// 	build := runtime.NewHostModuleBuilder(moduleName)
// 	for _, f := range funcs {
// 		build.NewFunctionBuilder().
// 			WithGoModuleFunction(f.f, f.input, f.output).
// 			WithName(f.name).
// 			Export(f.name)
// 	}
// 	return build.Compile(ctx)
// }
