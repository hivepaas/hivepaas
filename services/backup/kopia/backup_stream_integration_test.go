package kopia

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// A stream is taken as one file of a snapshot recorded under the source it is
// given, with its tags and description.
func TestIntegration_BackupStream_TakesTheStreamAsAFile(t *testing.T) {
	client, _ := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, &backupmodel.InitRepoOptions{}))

	resp, err := client.BackupStream(ctx, strings.NewReader("CREATE TABLE t();\n"), "db.sql",
		&backupmodel.BackupOptions{
			Source:      "hivepaas@data-backup:/j1",
			Description: "nightly (run t1)",
			Tags:        []string{"hivepaas.job:j1", "hivepaas.run:t1"},
		})

	mustNoError(t, err)
	mustNotNil(t, resp.Item)
	assert.Equal(t, int64(len("CREATE TABLE t();\n")), resp.Item.SizeBytes)
	assert.Contains(t, resp.Item.Paths, "/j1")
	assert.Equal(t, "data-backup", resp.Item.Hostname)
	assert.Equal(t, "nightly (run t1)", resp.Item.Description)
	assert.Contains(t, resp.Item.Tags, "hivepaas.run:t1")
}

// A directory is recorded under the source it is given too: every run of a job
// under one source, whichever node read it.
func TestIntegration_BackupDirectory_UnderAGivenSource(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, &backupmodel.InitRepoOptions{}))
	dataDir := filepath.Join(baseDir, "data")
	mustNoError(t, os.MkdirAll(dataDir, 0o755))
	mustNoError(t, os.WriteFile(filepath.Join(dataDir, "f1.txt"), []byte("hello"), 0o600))

	resp, err := client.BackupDirectory(ctx, dataDir, &backupmodel.BackupOptions{
		Source: "hivepaas@data-backup:/j2", Description: "files (run t2)",
	})

	mustNoError(t, err)
	mustNotNil(t, resp.Item)
	assert.Equal(t, []string{"/j2"}, resp.Item.Paths)
	assert.Equal(t, int64(len("hello")), resp.Item.SizeBytes)
	assert.Equal(t, "files (run t2)", resp.Item.Description)
}
