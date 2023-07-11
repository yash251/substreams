package goja

import (
	"github.com/streamingfast/logging"
)

var zlog, tracer = logging.PackageLogger("goja-runtime", "github.com/streamingfast/substreams/wasm/goja")
