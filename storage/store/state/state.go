package state

import (
	"fmt"

	"github.com/streamingfast/substreams/utils"

	"github.com/streamingfast/substreams/block"
	"go.uber.org/zap/zapcore"
)

// ModuleStorageState contains all the file-related ranges of store snapshots
// we'll want to plan work for, and things that are already available.
type StoreStorageState struct {
	ModuleName         string
	ModuleInitialBlock uint64

	LastCompletedRange     *FullStoreFile        // Points to a complete .kv file, to initialize the store upon getting started.
	MissingCompletedRanges MissingFullStoreFiles // 0-10.kv, 0-30.kv -> missing 0-20.kv
	PartialsMissing        PartialStoreFiles
	PartialsPresent        PartialStoreFiles
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
	if completeSnapshot != nil && completeSnapshot.ExclusiveEndBlock <= modInitBlock {
		return nil, fmt.Errorf("cannot have saved last store before module's init block")
	}

	// todo: battle test this
	if completeSnapshot != nil {
		totalNumberOfCompletedRanges := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
		if totalNumberOfCompletedRanges != uint64(snapshots.Completes.Len()) {
			var missingFullStoreFiles block.Ranges
			missingRangesCounter := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
			for i := snapshots.Completes.Len() - 1; i >= 0; i-- {
				completedEndBlock := snapshots.Completes[i].ExclusiveEndBlock / storeSaveInterval
				if completedEndBlock == totalNumberOfCompletedRanges {
					totalNumberOfCompletedRanges--
					missingRangesCounter--
				} else {
					for j := missingRangesCounter; j != completedEndBlock; j-- {
						missingFullStoreFiles = append(missingFullStoreFiles, block.NewRange(modInitBlock, totalNumberOfCompletedRanges*storeSaveInterval))
						totalNumberOfCompletedRanges--
						missingRangesCounter--
					}
				}
			}
		}
	}

	parallelProcessStartBlock := modInitBlock
	if completeSnapshot != nil {
		parallelProcessStartBlock = completeSnapshot.ExclusiveEndBlock
		out.LastCompletedRange = block.NewRange(modInitBlock, completeSnapshot.ExclusiveEndBlock)

		if completeSnapshot.ExclusiveEndBlock == workUpToBlockNum {
			return
		}
	}

	for ptr := parallelProcessStartBlock; ptr < workUpToBlockNum; {
		end := utils.MinOf(ptr-ptr%storeSaveInterval+storeSaveInterval, workUpToBlockNum)
		newPartial := block.NewRange(ptr, end)
		if !snapshots.ContainsPartial(newPartial) {
			out.PartialsMissing = append(out.PartialsMissing, newPartial)
		} else {
			out.PartialsPresent = append(out.PartialsPresent, newPartial)
		}
		ptr = end
	}
	return
}

func (s *StoreStorageState) Name() string { return s.ModuleName }

func (s *StoreStorageState) BatchRequests(subreqSplitSize uint64) block.Ranges {
	return s.PartialsMissing.MergedBuckets(subreqSplitSize)
}

func (s *StoreStorageState) InitialProgressRanges() (out block.Ranges) {
	if s.LastCompletedRange != nil {
		out = append(out, s.LastCompletedRange)
	}
	out = append(out, s.PartialsPresent.Merged()...)
	return
}
func (s *StoreStorageState) ReadyUpToBlock() uint64 {
	if s.LastCompletedRange == nil {
		return s.ModuleInitialBlock
	}
	return s.LastCompletedRange.ExclusiveEndBlock
}

func (w *StoreStorageState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("store_name", w.ModuleName)
	enc.AddString("last_completed_range", w.LastCompletedRange.String())
	enc.AddString("missing_completed_ranges", w.MissingCompletedRanges.String())
	enc.AddInt("partial_missing", len(w.PartialsMissing))
	enc.AddInt("partial_present", len(w.PartialsPresent))
	return nil
}

func computeMissingRanges(storeSaveInterval uint64, modInitBlock uint64, completeSnapshot *block.Range, snapshots *storeSnapshots) MissingFullStoreFiles {
	var missingFullStoreFiles block.Ranges

	totalNumberOfCompletedRanges := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
	if totalNumberOfCompletedRanges != uint64(snapshots.Completes.Len()) {
		missingRangesCounter := completeSnapshot.ExclusiveEndBlock / storeSaveInterval
		lastCompletedRange := totalNumberOfCompletedRanges
		for i := snapshots.Completes.Len() - 1; i >= 0; i-- {
			completedEndBlock := snapshots.Completes[i].ExclusiveEndBlock / storeSaveInterval
			if completedEndBlock == totalNumberOfCompletedRanges {
				totalNumberOfCompletedRanges--
				missingRangesCounter--
				lastCompletedRange--
			} else {
				for j := missingRangesCounter; j != completedEndBlock; j-- {
					missingFullStoreFiles = append(missingFullStoreFiles, block.NewRange(modInitBlock, totalNumberOfCompletedRanges*storeSaveInterval))
					totalNumberOfCompletedRanges--
					missingRangesCounter--
				}
			}
		}
	}

	return missingFullStoreFiles
}
