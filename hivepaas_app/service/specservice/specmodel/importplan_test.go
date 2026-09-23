package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectionSelects(t *testing.T) {
	const (
		global      = "global"
		projA       = "projects/a"
		projASet    = "projects/a/settings"
		dev         = "projects/a/envs/dev"
		devSettings = "projects/a/envs/dev/settings"
		backend     = "projects/a/envs/dev/apps/backend"
		frontend    = "projects/a/envs/dev/apps/frontend"
		prod        = "projects/a/envs/prod"
		projC       = "projects/c"
	)
	cases := map[string]struct {
		selection Selection
		selected  []string
		not       []string
	}{
		"global, every project": {Selection{}, []string{global, projA, backend, projC}, nil},
		"global, chosen projects": {
			Selection{Include: []string{global, projA}},
			[]string{global, projA, projASet, dev, backend},
			[]string{projC},
		},
		"project, chosen envs": {
			Selection{Include: []string{projASet, dev}},
			[]string{projASet, dev, devSettings, backend},
			[]string{projA, prod, global},
		},
		"env, chosen apps": {
			Selection{Include: []string{devSettings, backend}},
			[]string{devSettings, backend},
			[]string{dev, frontend},
		},
		"a wildcard segment": {
			Selection{Include: []string{"projects/*/settings"}},
			[]string{projASet, "projects/c/settings"},
			[]string{projA, dev},
		},
		"exclude wins": {
			Selection{Include: []string{projA}, Exclude: []string{"projects/a/envs/*"}},
			[]string{projA, projASet},
			[]string{dev, backend, prod},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for _, path := range tc.selected {
				assert.True(t, tc.selection.Selects(path), "%s should be selected", path)
			}
			for _, path := range tc.not {
				assert.False(t, tc.selection.Selects(path), "%s should not be selected", path)
			}
		})
	}
}
