package state

import (
	"fmt"

	"github.com/streamingfast/substreams/block"
	"github.com/streamingfast/substreams/storage/store"
	"github.com/streamingfast/substreams/utils"
	"go.uber.org/zap/zapcore"
)

// ModuleStorageState contains all the file-related ranges of store snapshots
// we'll want to plan work for, and things that are already available.
type StoreStorageState struct {
	ModuleName         string
	ModuleInitialBlock uint64

	InitialCompleteFile *store.FileInfo // Points to a complete .kv file, to initialize the store upon getting started.
	MissingCompletedRanges MissingFullStoreFiles // 0-10.kv, 0-30.kv -> missing 0-20.kv
	PartialsMissing     block.Ranges
}

type FullStoreFile = block.Range
type MissingFullStoreFiles = block.Ranges
type PartialStoreFiles = block.Ranges

func NewStoreStorageState(modName string, storeSaveInterval, modInitBlock, workUpToBlockNum uint64, snapshots *storeSnapshots) (out *StoreStorageState, err error) {
	out = &StoreStorageState{ModuleName: modName, ModuleInitialBlock: modInitBlock}
	if workUpToBlockNum <= modInitBlock {
		return
	}

	completeSnapshot := snapshots.LastCompleteSnapshotBefore(workUpToBlockNum)
	if completeSnapshot != nil && completeSnapshot.Range.ExclusiveEndBlock <= modInitBlock {
		return nil, fmt.Errorf("cannot have saved last store before module's init block")
	}

	if completeSnapshot != nil {
		out.MissingCompletedRanges = computeMissingRanges(storeSaveInterval, modInitBlock, completeSnapshot, snapshots)
	}

	parallelProcessStartBlock := modInitBlock
	if completeSnapshot != nil {
		parallelProcessStartBlock = completeSnapshot.Range.ExclusiveEndBlock
		out.InitialCompleteFile = completeSnapshot

		if completeSnapshot.Range.ExclusiveEndBlock == workUpToBlockNum {
			return
		}
	}

	for ptr := parallelProcessStartBlock; ptr < workUpToBlockNum; {
		end := utils.MinOf(ptr-ptr%storeSaveInterval+storeSaveInterval, workUpToBlockNum)
		out.PartialsMissing = append(out.PartialsMissing, block.NewRange(ptr, end))

		ptr = end
	}
	return
}

func (s *StoreStorageState) Name() string { return s.ModuleName }

func (s *StoreStorageState) BatchRequests(subreqSplitSize uint64) block.Ranges {
	return s.PartialsMissing.MergedBuckets(subreqSplitSize)
}

func (s *StoreStorageState) InitialProgressRanges() (out block.Ranges) {
	if s.InitialCompleteFile != nil {
		out = append(out, s.InitialCompleteFile.Range)
	}

	return
}
func (s *StoreStorageState) ReadyUpToBlock() uint64 {
	if s.InitialCompleteFile == nil {
		return s.ModuleInitialBlock
	}
	return s.InitialCompleteFile.Range.ExclusiveEndBlock
}

func (w *StoreStorageState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("store_name", w.ModuleName)
	bRange := "None"
	if w.InitialCompleteFile != nil {
		bRange = w.InitialCompleteFile.Range.String()
	}
	enc.AddString("intial_range", bRange)
	enc.AddInt("partial_missing", len(w.PartialsMissing))
	return nil
}

func computeMissingRanges(storeSaveInterval uint64, modInitBlock uint64, completeSnapshot *block.Range, snapshots *storeSnapshots) MissingFullStoreFiles {
	var missingFullStoreFiles block.Ranges

	totalNumberOfCompletedRanges := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
	if totalNumberOfCompletedRanges != uint64(snapshots.Completes.Len()) {
		missingRangesCounter := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
		for i := snapshots.Completes.Len() - 1; i >= 0; i-- {
			completedEndBlock := snapshots.Completes[i].ExclusiveEndBlock / storeSaveInterval
			if completedEndBlock == totalNumberOfCompletedRanges {
				totalNumberOfCompletedRanges--
				missingRangesCounter--
			} else {
				for j := missingRangesCounter; j >= completedEndBlock; j-- {
					if missingRangesCounter != completedEndBlock {
						missingFullStoreFiles = append(missingFullStoreFiles, block.NewRange(modInitBlock, totalNumberOfCompletedRanges*storeSaveInterval))
					}
					totalNumberOfCompletedRanges--
					missingRangesCounter--
				}
			}
		}
	}

	return missingFullStoreFiles
}
