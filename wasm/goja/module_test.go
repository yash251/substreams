package goja

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"

	"github.com/dop251/goja"
	"github.com/streamingfast/substreams/wasm"
	"github.com/stretchr/testify/require"
)

func TestRequireTextDecoder(t *testing.T) {
	tests := []struct {
		name     string
		codePath string
	}{
		{"text_decoder", "testdata/text_decoder.js"},
		{"text_encoder", "testdata/text_encoder.js"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			registry := wasm.NewRegistryWithRuntime("goja", nil, math.MaxUint64)

			module, err := newModule(ctx, readCode(t, tt.codePath), registry)
			require.NoError(t, err)

			_, err = module.ExecuteNewCall(ctx, wasm.NewCall(nil, "test", "run", []wasm.Argument{}), nil, []wasm.Argument{})
			if err != nil {
				var ex *goja.Exception
				if errors.As(err, &ex) {
					require.NoError(t, err, ex.String())
				}

				require.NoError(t, err)
			}
		})
	}
}

func readCode(t *testing.T, filename string) []byte {
	t.Helper()

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	return content
}
