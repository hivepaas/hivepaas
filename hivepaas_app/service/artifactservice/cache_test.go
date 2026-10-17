package artifactservice

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPruneCacheRemovesWhatNothingKeepsOnceItIsOld(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	old, recent := now.Add(-48*time.Hour), now.Add(-time.Hour)
	for name, modified := range map[string]time.Time{
		"kept-old":                  old,
		"unkept-old":                old,
		"unkept-recent":             recent, // downloaded for an app still being created
		".download-unkept-old-1234": old,    // left by an agent killed mid-download
	} {
		path := filepath.Join(dir, name)
		assert.NoError(t, os.WriteFile(path, []byte("12345"), 0o755))
		assert.NoError(t, os.Chtimes(path, modified, modified))
	}

	deleted, reclaimed, err := PruneCache(dir, []string{"kept-old"}, now.Add(-24*time.Hour))

	assert.NoError(t, err)
	assert.Equal(t, 2, deleted)
	assert.Equal(t, uint64(10), reclaimed)
	left, _ := os.ReadDir(dir)
	names := []string{}
	for _, entry := range left {
		names = append(names, entry.Name())
	}
	assert.ElementsMatch(t, []string{"kept-old", "unkept-recent"}, names)
}

func TestPruneCacheOfANodeWithoutOne(t *testing.T) {
	deleted, _, err := PruneCache(filepath.Join(t.TempDir(), "none"), nil, time.Now())

	assert.NoError(t, err)
	assert.Zero(t, deleted)
}
