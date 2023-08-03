package main

import (
	"fmt"
	"os"

	"context"
	"testing"

	"github.com/streamingfast/logging"
	"github.com/streamingfast/substreams/wasm"
	_ "github.com/streamingfast/substreams/wasm/wasmtime"
	_ "github.com/streamingfast/substreams/wasm/wazero"
	"github.com/stretchr/testify/require"
)

func init() {
	logging.InstantiateLoggers()
}

func BenchmarkExecution(b *testing.B) {
	type runtime struct {
		name                string
		code                []byte
		shouldReUseInstance bool
	}

	type testCase struct {
		tag        string
		entrypoint string
		arguments  []wasm.Argument
		// Right now there is differences between runtime, so we accept all those values
		acceptedByteCount []int
	}

	for _, testCase := range []*testCase{
		//{"bare", "map_noop", args(wasm.NewParamsInput("")), []int{0}},
		//
		//// Decode proto only decode and returns the block.number as the output (to ensure the block is not elided at compile time)
		//{"decode_proto_only", "map_decode_proto_only", args(blockInputFile(b, "testdata/ethereum_mainnet_block_16021772.binpb", "sf.ethereum.type.v2.Block")), []int{0}},
		//
		//{"map_block", "map_block", args(blockInputFile(b, "testdata/ethereum_mainnet_block_16021772.binpb", "sf.ethereum.type.v2.Block")), []int{44957, 45081}},

		{"map_sol_block_decoding", "map_sol_block_decoding", args(blockInputFile(b, "testdata/solana_mainnet_block_180279461.binpb", "sf.solana.type.v1.Block")), []int{0}},
		{"map_sol_block", "map_sol_block", args(blockInputFile(b, "testdata/solana_mainnet_block_180279461.binpb", "sf.solana.type.v1.Block")), []int{208395}},
		{"map_sol_block_owned", "map_sol_block_owned", args(blockInputFile(b, "testdata/solana_mainnet_block_180279461.binpb", "sf.solana.type.v1.Block")), []int{208395}},
	} {
		var reuseInstance = true
		var freshInstanceEachRun = false

		wasmCode := readCode(b, "substreams_wasm/substreams.wasm")

		for _, config := range []*runtime{
			{"wasmtime", wasmCode, reuseInstance},
			{"wasmtime", wasmCode, freshInstanceEachRun},

			{"wazero", wasmCode, reuseInstance},
			{"wazero", wasmCode, freshInstanceEachRun},
		} {
			instanceKey := "reused"
			if !config.shouldReUseInstance {
				instanceKey = "fresh"
			}

			b.Run(fmt.Sprintf("vm=%s,instance=%s,tag=%s", config.name, instanceKey, testCase.tag), func(b *testing.B) {
				ctx := context.Background()

				wasmRuntime := wasm.NewRegistryWithRuntime(config.name, nil, 0)

				module, err := wasmRuntime.NewModule(ctx, config.code)
				require.NoError(b, err)

				cachedInstance, err := module.NewInstance(ctx)
				require.NoError(b, err)
				defer cachedInstance.Close(ctx)

				call := wasm.NewCall(nil, testCase.tag, testCase.entrypoint, testCase.arguments)

				for i := 0; i < b.N; i++ {
					instance := cachedInstance
					if !config.shouldReUseInstance {
						instance, err = module.NewInstance(ctx)
						require.NoError(b, err)
					}

					_, err := module.ExecuteNewCall(ctx, call, instance, testCase.arguments)
					if err != nil {
						require.NoError(b, err)
					}

					require.Contains(b, testCase.acceptedByteCount, len(call.Output()), "invalid byte count got %d expected one of %v", len(call.Output()), testCase.acceptedByteCount)
				}
			})
		}
	}
}

func readCode(t require.TestingT, filename string) []byte {
	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	return content
}

func args(ins ...wasm.Argument) []wasm.Argument {
	return ins
}

func blockInputFile(t require.TestingT, filename string, sourceInput string) wasm.Argument {
	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	input := wasm.NewSourceInput(sourceInput)
	input.SetValue(content)

	return input
}
