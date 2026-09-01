package kopia

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func TestClient_ExecCommand_UsesCommandExecutor(t *testing.T) {
	var executedReq *backupmodel.CommandExecReq
	mockExecutor := func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		executedReq = req
		return &backupmodel.CommandExecResp{ExitCode: 0}, nil
	}

	c := NewClient(&backupmodel.Storage{
		RepositoryPassword: "secret-password",
		StorageLocal: &backupmodel.StorageLocal{
			Path:      "/mnt/backups",
			NodeID:    "node-1",
			NodeLabel: "backup-worker",
		},
	}, mockExecutor)

	err := c.CheckRepo(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, executedReq)
	assert.Equal(t, []string{"kopia", "repository", "validate-provider"}, executedReq.Command)
	assert.Equal(t, "node-1", executedReq.NodeID)
	assert.Equal(t, "backup-worker", executedReq.NodeLabel)
	assert.Contains(t, executedReq.Env, "KOPIA_PASSWORD=secret-password")

	// Test InitRepo
	err = c.InitRepo(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"kopia", "repository", "create", "filesystem", "--path=/mnt/backups",
		"--no-check-for-updates"}, executedReq.Command)

	// Test ConnectRepo
	err = c.ConnectRepo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"kopia", "repository", "connect", "filesystem", "--path=/mnt/backups",
		"--no-check-for-updates"}, executedReq.Command)
}

func TestClient_InitRepo_AppliesOptions(t *testing.T) {
	var executedCommands [][]string
	mockExecutor := func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		executedCommands = append(executedCommands, req.Command)
		return &backupmodel.CommandExecResp{ExitCode: 0}, nil
	}

	c := NewClient(&backupmodel.Storage{
		RepositoryPassword: "secret-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: "/mnt/backups"},
	}, mockExecutor)

	err := c.InitRepo(context.Background(), &backupmodel.InitRepoOptions{
		Description: "my repo",
		PackSizeMB:  32,
		Compression: "zstd-fastest",
	})
	assert.NoError(t, err)
	assert.Len(t, executedCommands, 3)
	assert.Equal(t, []string{"kopia", "repository", "create", "filesystem", "--path=/mnt/backups",
		"--no-check-for-updates", "--description=my repo"}, executedCommands[0])
	assert.Equal(t, []string{"kopia", "repository", "set-parameters", "--max-pack-size-mb=32"}, executedCommands[1])
	assert.Equal(t, []string{"kopia", "policy", "set", "--global", "--compression=zstd-fastest"},
		executedCommands[2])
}

// Each repository must keep its connection state in its own config file, otherwise operations on
// one repository act on whichever repository connected last.
func TestClient_ConfigFile_IsolatesRepos(t *testing.T) {
	var executedReq *backupmodel.CommandExecReq
	mockExecutor := func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		executedReq = req
		return &backupmodel.CommandExecResp{ExitCode: 0}, nil
	}

	c := NewClient(&backupmodel.Storage{
		RepositoryPassword: "secret-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: "/mnt/backups"},
		ConfigFile:         "/tmp/hivepaas/backup-repos/repo-1/repository.config",
	}, mockExecutor)

	err := c.CheckRepo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"kopia", "--config-file=/tmp/hivepaas/backup-repos/repo-1/repository.config",
		"repository", "validate-provider"}, executedReq.Command)
}
