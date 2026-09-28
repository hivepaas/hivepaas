package kopia

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// A stream comes back as it went in.
func TestIntegration_RestoreStream_GivesBackTheFile(t *testing.T) {
	client, _ := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, &backupmodel.InitRepoOptions{}))
	dump := "CREATE TABLE t();\nINSERT INTO t VALUES ();\n"
	taken, err := client.BackupStream(ctx, strings.NewReader(dump), "db.sql",
		&backupmodel.BackupOptions{Source: "hivepaas@data-backup:/j1"})
	mustNoError(t, err)

	var out bytes.Buffer
	_, err = client.RestoreStream(ctx, taken.Item.ID, "db.sql", &out, nil)

	mustNoError(t, err)
	assert.Equal(t, dump, out.String())
}

// What a snapshot holds is listed a directory at a time, and one directory of it
// is restored on its own.
func TestIntegration_ListEntriesThenRestoreOneDirectory(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, &backupmodel.InitRepoOptions{}))
	dataDir := filepath.Join(baseDir, "data")
	mustNoError(t, os.MkdirAll(filepath.Join(dataDir, "uploads", "2026"), 0o755))
	mustNoError(t, os.WriteFile(filepath.Join(dataDir, "app.db"), []byte("db"), 0o600))
	mustNoError(t, os.WriteFile(filepath.Join(dataDir, "uploads", "my photo.png"), []byte("png!"), 0o600))
	mustNoError(t, os.WriteFile(filepath.Join(dataDir, "uploads", "2026", "a.txt"), []byte("a"), 0o600))
	taken, err := client.BackupDirectory(ctx, dataDir, &backupmodel.BackupOptions{Source: "hivepaas@data-backup:/j2"})
	mustNoError(t, err)

	root, err := client.ListEntries(ctx, taken.Item.ID, "")
	mustNoError(t, err)
	assert.Equal(t, []backupmodel.SnapshotEntry{
		{Name: "uploads", Dir: true, SizeBytes: 5},
		{Name: "app.db", SizeBytes: 2},
	}, root)
	uploads, err := client.ListEntries(ctx, taken.Item.ID, "uploads")
	mustNoError(t, err)
	assert.Equal(t, []backupmodel.SnapshotEntry{
		{Name: "2026", Dir: true, SizeBytes: 1},
		{Name: "my photo.png", SizeBytes: 4},
	}, uploads)

	target := filepath.Join(baseDir, "restored", "uploads")
	_, err = client.RestoreDirectory(ctx, taken.Item.ID, target, &backupmodel.RestoreOptions{Path: "uploads"})

	mustNoError(t, err)
	got, err := os.ReadFile(filepath.Join(target, "2026", "a.txt"))
	mustNoError(t, err)
	assert.Equal(t, "a", string(got))
	_, err = os.Stat(filepath.Join(target, "app.db"))
	assert.True(t, os.IsNotExist(err), "only the directory asked for is restored")
}
