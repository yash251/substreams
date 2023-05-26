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

	InitialCompleteFile        *store.FileInfo
	MissingCompleteBlockRanges block.Ranges
	PartialsMissing            block.Ranges
}

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
		out.MissingCompleteBlockRanges = computeMissingRanges(storeSaveInterval, modInitBlock, completeSnapshot.Range, snapshots)
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
	// todo: add the missing full kv ranges here
	partialsMissingRequests := s.PartialsMissing.MergedBuckets(subreqSplitSize)
	missingCompleteBlockRangesRequests := s.MissingCompleteBlockRanges.MergedBuckets(subreqSplitSize)
	return partialsMissingRequests.Add(missingCompleteBlockRangesRequests)
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
	enc.AddString("initial_range", bRange)
	enc.AddInt("partial_missing", len(w.PartialsMissing))
	enc.AddString("missing_block_ranges", w.MissingCompleteBlockRanges.String())
	return nil
}

// ComputeMissingRanges will check the complete block ranges on disk and find the missing full kvs
func computeMissingRanges(storeSaveInterval uint64, modInitBlock uint64, completeSnapshot *block.Range, snapshots *storeSnapshots) block.Ranges {
	var missingFullStoreBlockRanges block.Ranges

	numberOfCompletedRanges := completeSnapshot.ExclusiveEndBlock / storeSaveInterval

	if numberOfCompletedRanges != uint64(snapshots.Completes.Ranges().Len()) {
		missingRangesCounter := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
		numberOfCompletedRangesFiles := snapshots.Completes.Ranges().Len()

		for i := numberOfCompletedRangesFiles - 1; i >= 0; i-- {
			lastCompletedFileEndRange := snapshots.Completes[i].Range.ExclusiveEndBlock / storeSaveInterval

			if lastCompletedFileEndRange == numberOfCompletedRanges {
				numberOfCompletedRanges--
				missingRangesCounter--
			} else {
				for j := missingRangesCounter; j >= lastCompletedFileEndRange; j-- {
					// [0-10, 0-20, 0-30, 0-60]
					// lastCompletedFileEndRange 30 -> totalNumberOfCompletedRanges -> 50
					// => add 50, 40 BUT not 30 as it's there
					if missingRangesCounter != lastCompletedFileEndRange {
						missingFullStoreBlockRanges = append(missingFullStoreBlockRanges, block.NewRange(modInitBlock, numberOfCompletedRanges*storeSaveInterval))
					}
					numberOfCompletedRanges--
					missingRangesCounter--
				}
			}
		}
	}

	return missingFullStoreBlockRanges
}
