package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeStatsOf(t *testing.T, stats string) *IndexEntry {
	t.Helper()
	index, err := DecodeIndex([]byte(`{
		"apiVersion": "hivepaas.com/v1", "kind": "TemplateIndex",
		"templates": [{"name": "demo", "file": {"path": "templates/demo.yaml", "sha256": "aa"},
			"versions": [{"name": "1"}], "stats": ` + stats + `}]
	}`))
	assert.NoError(t, err)
	assert.Empty(t, index.Skipped)
	return index.FindTemplate("demo")
}

func TestDecodeIndexReadsStats(t *testing.T) {
	entry := decodeStatsOf(t, `{"added": "2026-10-07", "stars": 7812, "starsGained": 340}`)

	assert.Equal(t, &IndexStats{Added: "2026-10-07", Stars: 7812, StarsGained: 340}, entry.Stats)
}

// The stats only order the store: one this HivePaaS cannot read is a template
// sorted as one without it, not a template left out.
func TestDecodeIndexDoesWithoutStatsItCannotRead(t *testing.T) {
	cases := map[string]struct {
		stats string
		want  *IndexStats
	}{
		"null":             {`null`, nil},
		"not an object":    {`"lots"`, &IndexStats{}},
		"a date it cannot": {`{"added": "October", "stars": 5}`, &IndexStats{Stars: 5}},
		"a changed type": {
			`{"added": "2026-10-07", "stars": {"count": 5}, "starsGained": 2}`,
			&IndexStats{Added: "2026-10-07", StarsGained: 2},
		},
		"a negative count":        {`{"stars": -1, "starsGained": 1.5}`, &IndexStats{}},
		"fields it does not know": {`{"stars": 5, "installs": {"week": 3}}`, &IndexStats{Stars: 5}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, decodeStatsOf(t, tc.stats).Stats)
		})
	}
}
