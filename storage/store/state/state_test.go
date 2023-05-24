package state

import (
	"strings"
	"testing"

	"github.com/streamingfast/substreams/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkUnits_init(t *testing.T) {
	type splitTestCase struct {
		name string

		modInitBlock uint64          // ModuleInitialBlock
		snapshots    *storeSnapshots // store's Last block saved from the store's Info file
		reqStart     uint64          // the request's absolute start block

		expectInitialCompletedRanges *block.Range // Used for LoadFrom()
		expectMissingPartialRanges   block.Ranges // sent to the user as already processed, and passed to the Squasher, the first Covered is expected to match the expectStoreInit
		expectPresentPartialRanges   block.Ranges // sent to the user as already processed, and passed to the Squasher, the first Covered is expected to match the expectStoreInit
		storeSaveInterval            uint64
	}

	splitTest := func(name string, storeSaveInterval uint64, modInitBlock uint64, snapshotsSpec string, reqStart uint64, expectInitialCompletedRanges, expectMissingPartialRanges, expectPresentPartialRanges string,
	) splitTestCase {
		c := splitTestCase{
			name:              name,
			storeSaveInterval: storeSaveInterval,
			snapshots:         parseSnapshotSpec(snapshotsSpec),
			modInitBlock:      modInitBlock,
			reqStart:          reqStart,
		}
		c.expectInitialCompletedRanges = block.ParseRange(expectInitialCompletedRanges)
		c.expectMissingPartialRanges = block.ParseRanges(expectMissingPartialRanges)
		c.expectPresentPartialRanges = block.ParseRanges(expectPresentPartialRanges)
		return c
	}

	for _, tt := range []splitTestCase{
		splitTest("simple", 10,
			/* modInit, snapshots, reqStart */
			50, "", 100,
			/* expected: initial progress, covered ranges, partials missing, present */
			"", "50-60, 60-70, 70-80, 80-90, 90-100", "",
		),
		splitTest("simple in-bound of interval", 10,
			1, "", 11,
			"", "1-10,10-11", "",
		),
		splitTest("nothing to work for, nothing to initialize", 10,
			55, "", 55,
			"", "", "",
		),
		splitTest("reqStart before module init, don't process anything and start with a clean store", 10,
			50, "", 10,
			"", "", "",
		),
		splitTest("one case", 10,
			0, "0-20,p20-30", 20,
			"0-20", "", "",
		),
		splitTest("10 blocks already processed", 10,
			50, "50-60,p70-80", 90,
			"50-60", "60-70,80-90", "70-80",
		),
		splitTest("40 blocks already processed", 10,
			50, "50-60,p60-70,p70-80", 100,
			"50-60", "80-90,90-100", "60-70,70-80",
		),
		splitTest("multiple complete", 10,
			50, "50-60,50-70,50-80,p80-90", 100, // would they be sorted this way? should we run `sort` on the snapshots first?
			"50-80", "90-100", "80-90",
		),
		splitTest("off bounds, no blocks processed", 10,
			55, "", 92,
			"", "55-60,60-70,70-80,80-90,90-92", "",
		),
		splitTest("reqStart just above the modInit, and lower bound lower than modInit", 10,
			55, "", 60,
			"", "55-60", "",
		),
		splitTest("reqStart just above the modInit, and lower bound lower than modInit, off bound", 10,
			55, "", 59,
			"", "55-59", "",
		),
		splitTest("reqStart just above the modInit, and lower bound lower than modInit, lastBlock higher", 10,
			55, "55-60,p60-70,p70-80", 60,
			"55-60", "", "",
		),
		splitTest("reqStart off bound just above the modInit, and lower bound lower than modInit, lastBlock higher", 10,
			55, "55-60,p60-70", 59,
			"", "55-59", "",
		),
		splitTest("reqStart equal to lastSaved, on bound", 10,
			50, "50-60,p60-70,p70-80,p80-90", 90,
			"50-60", "", "60-70,70-80,80-90",
		),
		splitTest("reqStart equal to lastSaved, off bound", 10,
			50, "50-60,p60-70,p70-80,p80-90", 92,
			"50-60", "90-92", "60-70,70-80,80-90",
		),
	} {
		t.Run(tt.name, func(t *testing.T) {
			wu, err := NewStoreStorageState("mod", tt.storeSaveInterval, tt.modInitBlock, tt.reqStart, tt.snapshots)
			require.NoError(t, err)
			assert.Equal(t, tt.expectInitialCompletedRanges, wu.LastCompletedRange)
			assert.Equal(t,
				tt.expectMissingPartialRanges.String(),
				wu.PartialsMissing.String(),
			)
			assert.Equal(t,
				tt.expectPresentPartialRanges.String(),
				wu.PartialsPresent.String(),
			)
		})
	}
}

func Test_MissingParallelKvs(t *testing.T) {
	tests := []struct {
		name              string
		modName           string
		saveStoreInterval uint64
		modInitBlock      uint64
		workUpToBlock     uint64
		snapshots         *storeSnapshots

		expectedLastCompletedRange     *FullStoreFile
		expectedMissingCompletedRanges MissingFullStoreFiles
		expectedPartialsMissing        PartialStoreFiles
		expectedPartialsPresent        PartialStoreFiles
	}{
		{
			name:              "simple job, nothing missing",
			modName:           "module.a",
			saveStoreInterval: 10,
			modInitBlock:      0,
			workUpToBlock:     10,
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-10"),
				Partials:  nil,
			},
			expectedLastCompletedRange: &FullStoreFile{
				StartBlock:        0,
				ExclusiveEndBlock: 10,
			},
			expectedMissingCompletedRanges: nil,
			expectedPartialsMissing:        nil,
			expectedPartialsPresent:        nil,
		},
		{
			name:              "simple job, some partials missing",
			modName:           "module.a",
			saveStoreInterval: 10,
			modInitBlock:      0,
			workUpToBlock:     40,
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-10"),
				Partials:  block.ParseRanges("10-20,20-30"),
			},
			expectedLastCompletedRange: &FullStoreFile{
				StartBlock:        0,
				ExclusiveEndBlock: 10,
			},
			expectedMissingCompletedRanges: nil,
			expectedPartialsMissing:        block.ParseRanges("30-40"),
			expectedPartialsPresent:        block.ParseRanges("10-20,20-30"),
		},
		{
			name:              "simple job with more completed ranges, no holes and no partials",
			modName:           "module.a",
			saveStoreInterval: 10,
			modInitBlock:      0,
			workUpToBlock:     100,
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-10,10-20,20-30,30-40"),
				Partials:  nil,
			},
			expectedLastCompletedRange: &FullStoreFile{
				StartBlock:        0,
				ExclusiveEndBlock: 40,
			},
			expectedMissingCompletedRanges: nil,
			expectedPartialsMissing:        block.ParseRanges("30-40"),
			expectedPartialsPresent:        block.ParseRanges("10-20,20-30"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storeStorageState, err := NewStoreStorageState(test.modName, test.saveStoreInterval, test.modInitBlock, test.workUpToBlock, test.snapshots)
			require.NoError(t, err)

			assert.Equal(t, test.expectedLastCompletedRange, storeStorageState.LastCompletedRange)
			assert.Equal(t, test.expectedMissingCompletedRanges, storeStorageState.MissingCompletedRanges)
			assert.Equal(t, test.expectedPartialsMissing, storeStorageState.PartialsMissing)
			assert.Equal(t, test.expectedPartialsPresent, storeStorageState.PartialsPresent)
		})
	}
}

func Test_computeMissingRanges(t *testing.T) {
	tests := []struct {
		name              string
		storeSaveInterval uint64
		modInitBlock      uint64
		completeSnapshot  *block.Range
		snapshots         *storeSnapshots

		expectedMissingFullStoreFiles block.Ranges
	}{
		{
			name:              "missing 1 full store",
			storeSaveInterval: 100,
			modInitBlock:      0,
			completeSnapshot: &block.Range{
				StartBlock:        0,
				ExclusiveEndBlock: 500,
			},
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-100,0-200,0-300,0-500"),
				Partials:  nil,
			},
			expectedMissingFullStoreFiles: block.ParseRanges("0-400"),
		},
		{
			name:              "missing 2 full stores",
			storeSaveInterval: 100,
			modInitBlock:      0,
			completeSnapshot: &block.Range{
				StartBlock:        0,
				ExclusiveEndBlock: 500,
			},
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-100,0-200,0-500"),
				Partials:  nil,
			},
			expectedMissingFullStoreFiles: block.ParseRanges("0-400,0-300"),
		},
		{
			name:              "missing 2 full stores at different intervals",
			storeSaveInterval: 100,
			modInitBlock:      0,
			completeSnapshot: &block.Range{
				StartBlock:        0,
				ExclusiveEndBlock: 500,
			},
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-100,0-300,0-500"),
				Partials:  nil,
			},
			expectedMissingFullStoreFiles: block.ParseRanges("0-400,0-200"),
		},
		{
			name:              "missing multiple full stores at different intervals",
			storeSaveInterval: 100,
			modInitBlock:      0,
			completeSnapshot: &block.Range{
				StartBlock:        0,
				ExclusiveEndBlock: 1000,
			},
			snapshots: &storeSnapshots{
				Completes: block.ParseRanges("0-100,0-300,0-500,0-700,0-900"),
				Partials:  nil,
			},
			expectedMissingFullStoreFiles: block.ParseRanges("0-1000,0-800,0-600,0-400,0-200"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			missingRanges := computeMissingRanges(test.storeSaveInterval, test.modInitBlock, test.completeSnapshot, test.snapshots)
			assert.Equal(t, test.expectedMissingFullStoreFiles, missingRanges)
		})
	}
}

func parseSnapshotSpec(in string) *storeSnapshots {
	out := &storeSnapshots{}
	if in == "" {
		return out
	}
	for _, el := range strings.Split(in, ",") {
		el = strings.Trim(el, " ")
		partial := strings.Contains(el, "p")
		partRange := block.ParseRange(strings.Trim(el, "p"))
		if partial {
			out.Partials = append(out.Partials, partRange)
		} else {
			out.Completes = append(out.Completes, partRange)
		}
	}
	out.Sort()
	return out
}
