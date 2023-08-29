package config

import (
	"time"

	"github.com/streamingfast/dstore"

	"github.com/streamingfast/substreams/orchestrator/work"
)

// RuntimeConfig is a global configuration for the service.
// It is passed down and should not be modified unless cloned.
type RuntimeConfig struct {
	StateBundleSize uint64

	MaxWasmFuel        uint64        // if not 0, enable fuel consumption monitoring to stop runaway wasm module processing forever
	MaxJobsAhead       uint64        // limit execution of depencency jobs so they don't go too far ahead of the modules that depend on them (ex: module X is 2 million blocks ahead of module Y that depends on it, we don't want to schedule more module X jobs until Y caught up a little bit)
	InitSubrequests    int           // how many sub-jobs to start exactly when the tier1 request comes in
	SubrequestsRampup  time.Duration // over this amount of time, parallel jobs count will go from 'InitParallelSubrequests' to the 'ParallelSubrequests' (default or overriden per request)
	DefaultSubrequests int           // how many sub-jobs to launch for a given user
	// derives substores `states/`, for `store` modules snapshots (full and partial)
	// and `outputs/` for execution output of both `map` and `store` module kinds
	BaseObjectStore dstore.Store
	DefaultCacheTag string // appended to BaseObjectStore unless overriden by auth layer
	WorkerFactory   work.WorkerFactory

	ModuleExecutionTracing bool
}

func NewRuntimeConfig(
	stateBundleSize uint64,
	parallelSubrequests uint64,
	initParallelSubrequests uint64,
	parallelRampupPeriod time.Duration,
	maxJobsAhead uint64,
	maxWasmFuel uint64,
	baseObjectStore dstore.Store,
	defaultCacheTag string,
	workerFactory work.WorkerFactory,
) RuntimeConfig {
	return RuntimeConfig{
		StateBundleSize:    stateBundleSize,
		DefaultSubrequests: int(parallelSubrequests),
		InitSubrequests:    int(initParallelSubrequests),
		SubrequestsRampup:  parallelRampupPeriod,
		MaxJobsAhead:       maxJobsAhead,
		MaxWasmFuel:        maxWasmFuel,
		BaseObjectStore:    baseObjectStore,
		DefaultCacheTag:    defaultCacheTag,
		WorkerFactory:      workerFactory,
		// overridden by Tier Options
		ModuleExecutionTracing: false,
	}
}
