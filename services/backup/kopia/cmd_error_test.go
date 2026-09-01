package kopia

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func failingClient(t *testing.T) *Client {
	t.Helper()
	return NewClient(&backupmodel.Storage{
		RepositoryPassword: "super-secret-password",
		StorageS3: &backupmodel.StorageS3{
			Bucket:    "backups",
			AccessKey: "AKIAEXAMPLEKEY",
			SecretKey: "s3cr3t-access-key",
		},
	}, func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		return &backupmodel.CommandExecResp{ExitCode: 1}, nil
	})
}

// Command failures carry the command in their message, and that message travels to the API
// response. Credentials go through the environment, so they are simply not there to leak.
func TestExecCommand_ErrorCarriesNoCredentials(t *testing.T) {
	err := failingClient(t).InitRepo(context.Background(), nil)
	assert.Error(t, err)

	assert.NotContains(t, err.Error(), "s3cr3t-access-key")
	assert.NotContains(t, err.Error(), "AKIAEXAMPLEKEY")
	assert.NotContains(t, err.Error(), "super-secret-password")

	// The non-sensitive part must survive, otherwise the message is useless for debugging.
	assert.Contains(t, err.Error(), "--bucket=backups")
	assert.Contains(t, err.Error(), "repository create")
}

// redactCommand is a backstop for any future flag that does carry a secret. It is not the
// mechanism that keeps credentials safe today - the environment is.
func TestRedactCommand_MasksSecretFlags(t *testing.T) {
	got := redactCommand([]string{
		"kopia", "repository", "create", "s3",
		"--bucket=backups",
		"--access-key=AKIAEXAMPLEKEY",
		"--secret-access-key=s3cr3t",
		"--password=hunter2",
	})

	assert.Equal(t, "kopia repository create s3 --bucket=backups "+
		"--access-key=*** --secret-access-key=*** --password=***", got)
}

func TestExecCommand_FailureCarriesTypedError(t *testing.T) {
	err := failingClient(t).InitRepo(context.Background(), nil)
	assert.Error(t, err)

	assert.True(t, errors.Is(err, backupmodel.ErrCommandFailed))
	assert.True(t, errors.Is(err, hperrors.ErrPreconditionFailed))

	hpErr, ok := err.(hperrors.HPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusPreconditionFailed, hpErr.StatusCode())
}

func TestBuildStorageFlags_TypedErrors(t *testing.T) {
	noStorage := NewClient(nil, backupmodel.DefaultCommandExecutor)
	_, err := noStorage.buildStorageFlags()
	assert.True(t, errors.Is(err, backupmodel.ErrStorageConfigRequired))
	assert.True(t, errors.Is(err, hperrors.ErrBadRequest))

	emptyStorage := NewClient(&backupmodel.Storage{}, backupmodel.DefaultCommandExecutor)
	_, err = emptyStorage.buildStorageFlags()
	assert.True(t, errors.Is(err, backupmodel.ErrStorageTypeRequired))
}

func TestExecCommand_MissingExecutorIsTyped(t *testing.T) {
	c := NewClient(&backupmodel.Storage{
		StorageLocal: &backupmodel.StorageLocal{Path: "/tmp/repo"},
	}, nil)

	_, err := c.execCommand(context.Background(), []string{"repository", "status"})
	assert.True(t, errors.Is(err, backupmodel.ErrCommandExecutorMissing))
	assert.True(t, errors.Is(err, hperrors.ErrInternal))
}

func TestRedactArg(t *testing.T) {
	assert.Equal(t, "--bucket=backups", redactArg("--bucket=backups"))
	assert.Equal(t, "--secret-access-key=***", redactArg("--secret-access-key=abc"))
	assert.Equal(t, "--password=***", redactArg("--password=abc"))
	assert.Equal(t, "--new-password=***", redactArg("--new-password=abc"))
	assert.Equal(t, "repository", redactArg("repository"))
	// A value that merely contains the flag name must not be mistaken for it.
	assert.Equal(t, "--endpoint=--password", redactArg("--endpoint=--password"))
}
