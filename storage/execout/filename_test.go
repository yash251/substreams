package execout

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/streamingfast/substreams/block"
)

func Test_fileNameToRange(t *testing.T) {
	type args struct {
		filename string
		regex    *regexp.Regexp
	}
	tests := []struct {
		name    string
		args    args
		want    *block.Range
		wantErr bool
	}{
		{
			name: "cache filename",
			args: args{
				filename: "0013368000-0013369000.output",
				regex:    cacheFilenameRegex,
			},
			want: &block.Range{
				StartBlock:        13368000,
				ExclusiveEndBlock: 13369000,
			},
			wantErr: false,
		},
		{
			name: "index filename",
			args: args{
				filename: "0000122000-0000123000.index",
				regex:    indexFilenameRegex,
			},
			want: &block.Range{
				StartBlock:        122000,
				ExclusiveEndBlock: 123000,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fileNameToRange(tt.args.filename, tt.args.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("fileNameToRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fileNameToRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
