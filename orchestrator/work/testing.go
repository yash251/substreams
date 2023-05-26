package work

import (
	"strings"

	"github.com/streamingfast/substreams/storage/store"

	"go.uber.org/zap"

	state2 "github.com/streamingfast/substreams/storage/store/state"

	"github.com/streamingfast/substreams/storage/execout/state"

	"github.com/streamingfast/substreams/block"
	"github.com/streamingfast/substreams/storage"
)

func TestJob(modName string, rng string, prio int) *Job {
	return NewJob(modName, block.ParseRange(rng), nil, prio)
}

func TestPlanReadyJobs(jobs ...*Job) *Plan {
	return &Plan{
		readyJobs:                 jobs,
		highestModuleRunningBlock: map[string]uint64{},
		logger:                    zap.NewNop(),
	}
}

func TestJobDeps(modName string, rng string, prio int, deps string) *Job {
	return NewJob(modName, block.ParseRange(rng), strings.Split(deps, ","), prio)
}

func TestStoreStatePartialsMissing(modName string, rng string) storage.ModuleStorageState {
	return &state2.StoreStorageState{ModuleName: modName, PartialsMissing: block.ParseRanges(rng)}
}

func TestStoreStateCompleteRangesAndMissingRanges(modName string, completeFile *store.FileInfo, missingRanges string) storage.ModuleStorageState {
	return &state2.StoreStorageState{
		ModuleName:                 modName,
		InitialCompleteFile:        completeFile,
		MissingCompleteBlockRanges: block.ParseRanges(missingRanges),
	}
}

func TestStoreStateCompleteRangesMissingRangesPartialsMissing(modName string, completeFile *store.FileInfo, missingRanges string, partialRanges string) storage.ModuleStorageState {
	return &state2.StoreStorageState{
		ModuleName:                 modName,
		InitialCompleteFile:        completeFile,
		MissingCompleteBlockRanges: block.ParseRanges(missingRanges),
		PartialsMissing:            block.ParseRanges(partialRanges),
	}
}

func TestMapState(modName string, rng string) storage.ModuleStorageState {
	return &state.ExecOutputStorageState{ModuleName: modName, SegmentsMissing: block.ParseRanges(rng)}
}

func TestModStateMap(modStates ...storage.ModuleStorageState) (out storage.ModuleStorageStateMap) {
	out = make(storage.ModuleStorageStateMap)
	for _, mod := range modStates {
		out[mod.Name()] = mod
	}
	return
}
