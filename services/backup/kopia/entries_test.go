package kopia

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// `kopia ls -l` prints a line an entry: mode, size, time, object ID, name - a
// directory's with a trailing slash. They come back directories first, then by
// name.
func TestParseEntries(t *testing.T) {
	out := "drwxr-xr-x            7 2026-09-28 17:31:10 +07 k95041eface83bb618facd6308159580f  sub/\n" +
		"-rw-r--r--            2 2026-09-28 17:31:10 +07 bcb9b7c1cd57764aea2a28ac6731cfd2   a.txt\n" +
		"-rw-r--r--         1024 2026-09-28 17:31:10 UTC Ix0102   my photo.png\n" +
		"lrwxrwxrwx            4 2026-09-28 17:31:10 +07 ab12   link\n" +
		"\n"

	entries, err := parseEntries(out)

	assert.NoError(t, err)
	assert.Equal(t, []listedEntry{
		{SnapshotEntry: backupmodel.SnapshotEntry{Name: "sub", Dir: true, SizeBytes: 7},
			objectID: "k95041eface83bb618facd6308159580f"},
		{SnapshotEntry: backupmodel.SnapshotEntry{Name: "a.txt", SizeBytes: 2},
			objectID: "bcb9b7c1cd57764aea2a28ac6731cfd2"},
		{SnapshotEntry: backupmodel.SnapshotEntry{Name: "link", SizeBytes: 4}, objectID: "ab12"},
		{SnapshotEntry: backupmodel.SnapshotEntry{Name: "my photo.png", SizeBytes: 1024}, objectID: "Ix0102"},
	}, entries)
}

func TestParseEntriesRefusesWhatItCannotRead(t *testing.T) {
	_, err := parseEntries("something else\n")

	assert.Error(t, err)
}

// A path inside a snapshot is joined to its ID as kopia names it.
func TestSnapshotObjectPath(t *testing.T) {
	assert.Equal(t, "k1", snapshotObjectPath("k1", ""))
	assert.Equal(t, "k1/uploads/2026", snapshotObjectPath("k1", "uploads/2026"))
	assert.Equal(t, "k1/uploads", snapshotObjectPath("k1", "/uploads/"))
}
