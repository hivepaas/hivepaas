package reposerveragentuc

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
	"github.com/hivepaas/hivepaas/services/backup/kopia"
)

// A real kopia server run for a session takes a stream from a client logged in
// as the session's user, and is gone once the session ends.
func TestIntegration_RunServesARepository(t *testing.T) {
	if _, err := exec.LookPath("kopia"); err != nil {
		t.Skip("kopia binary not installed, skipping integration test")
	}
	repoDir := t.TempDir()
	owner := kopia.NewClient(&backupmodel.Storage{
		RepositoryPassword: "repo-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: repoDir},
		ConfigFile:         filepath.Join(t.TempDir(), "owner.config"),
	}, backupmodel.DefaultCommandExecutor)
	if err := owner.InitRepo(context.Background(), &backupmodel.InitRepoOptions{}); err != nil {
		t.Fatal(err)
	}

	uc := newTestUC(t, nil)
	uc.command = kopiaCommand
	uc.startTimeout = 30 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readyCh := make(chan *Ready, 1)
	done := make(chan error, 1)
	go func() {
		done <- uc.Run(ctx, &RunReq{
			RepoPath: repoDir, RepoPassword: "repo-password",
			Username: "hivepaas@data-backup", UserPassword: "user-password", ListenHost: "127.0.0.1",
		}, func(r *Ready) error {
			readyCh <- r
			return nil
		})
	}()

	var ready *Ready
	select {
	case ready = <-readyCh:
	case err := <-done:
		t.Fatalf("the server did not start: %v", err)
	case <-time.After(60 * time.Second):
		t.Fatal("the server did not start in time")
	}

	client := kopia.NewClient(&backupmodel.Storage{
		RepositoryPassword: "user-password",
		StorageServer: &backupmodel.StorageServer{
			URL: "https://127.0.0.1:" + strconv.Itoa(ready.Port), Fingerprint: ready.Fingerprint,
			Username: "hivepaas", Hostname: "data-backup",
		},
		ConfigFile: filepath.Join(t.TempDir(), "client.config"),
	}, backupmodel.DefaultCommandExecutor)
	assert.NoError(t, client.ConnectRepo(ctx))
	resp, err := client.BackupStream(ctx, strings.NewReader("CREATE TABLE t();\n"), "db.sql",
		&backupmodel.BackupOptions{Source: "hivepaas@data-backup:/j1"})
	if assert.NoError(t, err) && assert.NotNil(t, resp.Item) {
		assert.NotEmpty(t, resp.Item.ID)
	}

	cancel()
	assert.NoError(t, <-done)
	entries, _ := os.ReadDir(uc.sessionRoot)
	assert.Empty(t, entries, "the session directory is removed")
	out, _ := exec.Command("pgrep", "-f", uc.sessionRoot).Output()
	assert.Empty(t, strings.TrimSpace(string(out)), "no kopia server is left")
}
