package version

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *Version
		wantErr bool
	}{
		{
			name:    "full version with v prefix",
			version: "v1.2.3-beta02",
			want: &Version{
				Major:        1,
				Minor:        2,
				Patch:        3,
				Suffix:       "beta",
				SuffixNumber: 2,
			},
			wantErr: false,
		},
		{
			name:    "version without v prefix",
			version: "1.2.3-beta02",
			want: &Version{
				Major:        1,
				Minor:        2,
				Patch:        3,
				Suffix:       "beta",
				SuffixNumber: 2,
			},
			wantErr: false,
		},
		{
			name:    "version without suffix",
			version: "v1.2.3",
			want: &Version{
				Major:        1,
				Minor:        2,
				Patch:        3,
				Suffix:       "",
				SuffixNumber: 0,
			},
			wantErr: false,
		},
		{
			name:    "version with suffix but no suffix number",
			version: "1.2.3-rc",
			want: &Version{
				Major:        1,
				Minor:        2,
				Patch:        3,
				Suffix:       "rc",
				SuffixNumber: 0,
			},
			wantErr: false,
		},
		{
			name:    "invalid version",
			version: "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing patch",
			version: "v1.2-beta02",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCmp(t *testing.T) {
	tests := []struct {
		name string
		v1   *Version
		v2   *Version
		want int // <0 for v1 < v2, 0 for v1 == v2, >0 for v1 > v2
	}{
		{
			name: "equal versions",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta", SuffixNumber: 2},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta", SuffixNumber: 2},
			want: 0,
		},
		{
			name: "major less",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3},
			v2:   &Version{Major: 2, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "major greater",
			v1:   &Version{Major: 2, Minor: 0, Patch: 0},
			v2:   &Version{Major: 1, Minor: 9, Patch: 9},
			want: 1,
		},
		{
			name: "minor less",
			v1:   &Version{Major: 1, Minor: 1, Patch: 3},
			v2:   &Version{Major: 1, Minor: 2, Patch: 0},
			want: -1,
		},
		{
			name: "minor greater",
			v1:   &Version{Major: 1, Minor: 3, Patch: 0},
			v2:   &Version{Major: 1, Minor: 2, Patch: 9},
			want: 1,
		},
		{
			name: "patch less",
			v1:   &Version{Major: 1, Minor: 2, Patch: 2},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3},
			want: -1,
		},
		{
			name: "patch greater",
			v1:   &Version{Major: 1, Minor: 2, Patch: 4},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3},
			want: 1,
		},
		{
			name: "suffix less",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "alpha"},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta"},
			want: -1,
		},
		{
			name: "suffix greater",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "rc"},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta"},
			want: 1,
		},
		{
			name: "suffix number less",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta", SuffixNumber: 1},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta", SuffixNumber: 2},
			want: -1,
		},
		{
			name: "suffix number greater",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta", SuffixNumber: 3},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3, Suffix: "beta", SuffixNumber: 2},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cmp(tt.v1, tt.v2)
			if (got < 0 && tt.want >= 0) || (got > 0 && tt.want <= 0) || (got == 0 && tt.want != 0) {
				t.Errorf("Cmp() = %v, want %v (sign)", got, tt.want)
			}
		})
	}
}
