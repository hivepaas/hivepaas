package backupreposerviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// A command the engine gives no node runs on the node asked for: the engine of
// a cloud repository names none, and a directory on a node has to be read there.
func TestOnNodeExecutorRunsWhereAsked(t *testing.T) {
	var got *backupmodel.CommandExecReq
	base := func(_ context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		got = req
		return &backupmodel.CommandExecResp{}, nil
	}

	exec := onNodeExecutor(base, "node-2", "")
	_, err := exec(context.Background(), &backupmodel.CommandExecReq{Command: []string{"kopia"}})

	assert.NoError(t, err)
	assert.Equal(t, "node-2", got.NodeID)

	exec = onNodeExecutor(base, "", "zone=b")
	_, _ = exec(context.Background(), &backupmodel.CommandExecReq{Command: []string{"kopia"}})
	assert.Equal(t, "zone=b", got.NodeLabel)
}

// A repository on a volume is reachable on its own node only: a directory on
// another node cannot be backed up into it yet.
func TestCheckRepoReachableFrom(t *testing.T) {
	cloud := &backup.Storage{StorageS3: &backup.StorageS3{Bucket: "b"}}
	assert.NoError(t, checkRepoReachableFrom(cloud, "node-2", ""))

	onNode2 := &backup.Storage{StorageLocal: &backup.StorageLocal{Path: "/host/srv", NodeID: "node-2"}}
	assert.NoError(t, checkRepoReachableFrom(onNode2, "node-2", ""))
	err := checkRepoReachableFrom(onNode2, "node-3", "")
	assert.True(t, errors.Is(err, hperrors.ErrNotImplemented), "got %v", err)

	byLabel := &backup.Storage{StorageLocal: &backup.StorageLocal{Path: "/host/srv", NodeLabel: "zone=b"}}
	assert.NoError(t, checkRepoReachableFrom(byLabel, "", "zone=b"))
	assert.Error(t, checkRepoReachableFrom(byLabel, "node-2", ""))
}

// A snapshot the engine made reads as the repository's snapshots do.
func TestToRepoSnapshot(t *testing.T) {
	when := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	snapshot := toRepoSnapshot(&backupmodel.Snapshot{
		ID: "k1", ShortID: "k1s", Time: when, Tags: []string{"a:b"}, Paths: []string{"/db.sql"},
		Hostname: "h", SizeBytes: 42,
	})

	assert.Equal(t, &entity.BackupSnapshot{
		ID: "k1", ShortID: "k1s", Time: when, Paths: []string{"/db.sql"}, Hostname: "h", SizeBytes: 42,
	}, snapshot.Snapshot)
	assert.Equal(t, []string{"a:b"}, snapshot.Tags)
	assert.Nil(t, toRepoSnapshot(nil))
}
