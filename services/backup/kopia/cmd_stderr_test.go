package kopia

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

const lstatFailure = "unable to get local filesystem entry: resolveSymlink: stat: " +
	"lstat /host/srv/data: no such file or directory\n"

// clientPrinting fails its commands the way kopia does: a message on stderr,
// exit code 1.
func clientPrinting(stderr string) *Client {
	return NewClient(&backupmodel.Storage{
		RepositoryPassword: "p",
		StorageLocal:       &backupmodel.StorageLocal{Path: "/mnt/backups"},
	}, func(_ context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		if req.Stderr != nil {
			_, _ = io.WriteString(req.Stderr, stderr)
		}
		return &backupmodel.CommandExecResp{ExitCode: 1}, nil
	})
}

// A command that fails says why, in kopia's words: the error's detail - what a
// run's page and an API response show - carries what kopia printed.
func TestExecCommand_FailureDetailCarriesKopiasWords(t *testing.T) {
	_, err := clientPrinting(lstatFailure).BackupDirectory(context.Background(), "/host/srv/data", nil)

	assert.ErrorIs(t, err, backupmodel.ErrCommandFailed)
	detail := hperrors.GetErrorDetail(err, "")
	assert.Contains(t, detail, "ERR_BACKUP_COMMAND_FAILED")
	assert.Contains(t, detail, "lstat /host/srv/data: no such file or directory")
}

// Only the end of a long stderr is kept: the reason is at the end.
func TestExecCommand_FailureDetailKeepsTheEndOfALongStderr(t *testing.T) {
	long := strings.Repeat("progress line\n", 500) + "fatal: repository not found\n"

	_, err := clientPrinting(long).BackupDirectory(context.Background(), "/srv", nil)

	detail := hperrors.GetErrorDetail(err, "")
	assert.Contains(t, detail, "fatal: repository not found")
	assert.Less(t, len(detail), 3000)
}

// A caller that reads stderr itself still gets all of it; one that reads stdout
// only gets stdout alone - kopia's JSON, not its warnings - whichever node runs
// the command, and the reason still reaches the error.
func TestExecCommand_StderrStillReachesTheCaller(t *testing.T) {
	c := clientPrinting(lstatFailure)

	var stderr bytes.Buffer
	_, _ = c.execCommand(context.Background(), []string{"snapshot", "list"}, func(o *execOptions) {
		o.stderr = &stderr
	})
	assert.Equal(t, lstatFailure, stderr.String())

	var stdout bytes.Buffer
	_, err := c.execCommand(context.Background(), []string{"snapshot", "list"}, func(o *execOptions) {
		o.stdout = &stdout
	})
	assert.Empty(t, stdout.String())
	assert.Contains(t, hperrors.GetErrorDetail(err, ""), "no such file or directory")
}

// Deleting a snapshot the repository no longer holds says so, typed: the caller
// counts it as deleted.
func TestDeleteSnapshot_GoneIsNotFound(t *testing.T) {
	c := clientPrinting("ERROR error deleting snapshots by root ID k1: no snapshots matched k1\n")

	_, err := c.DeleteSnapshot(context.Background(), "k1")

	assert.ErrorIs(t, err, backupmodel.ErrSnapshotNotFound)
}
