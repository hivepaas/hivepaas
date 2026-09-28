package backupreposerviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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

// Where a backup runs: a repository on a volume is reached directly on its own
// node only; a stream into it, or a directory on another node, needs its server.
func TestRepoServerNeeded(t *testing.T) {
	cloud := &backup.Storage{StorageS3: &backup.StorageS3{Bucket: "b"}}
	onNode2 := &backup.Storage{StorageLocal: &backup.StorageLocal{Path: "/host/srv", NodeID: "node-2"}}
	byLabel := &backup.Storage{StorageLocal: &backup.StorageLocal{Path: "/host/srv", NodeLabel: "zone=b"}}
	shared := &backup.Storage{StorageLocal: &backup.StorageLocal{Path: "/host/nfs", NodeID: "node-1", Shared: true}}

	cases := []struct {
		name      string
		storage   *backup.Storage
		stream    bool
		nodeID    string
		nodeLabel string
		want      bool
	}{
		{"a stream into cloud storage", cloud, true, "", "", false},
		{"a directory into cloud storage", cloud, false, "node-3", "", false},
		{"a stream into a volume repository", onNode2, true, "", "", true},
		{"a directory on the repository's node", onNode2, false, "node-2", "", false},
		{"a directory on another node", onNode2, false, "node-3", "", true},
		{"a directory on the repository's labeled node", byLabel, false, "", "zone=b", false},
		{"a directory on a node without the label", byLabel, false, "node-2", "", true},
		{"a stream into a shared volume repository", shared, true, "", "", true},
		{"a directory on any node into a shared volume repository", shared, false, "node-3", "", false},
		{"a directory on a labeled node into a shared volume repository", shared, false, "", "zone=b", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, repoServerNeeded(c.storage, c.stream, c.nodeID, c.nodeLabel), c.name)
	}
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

// A directory backed up into a repository every node reaches is backed up where
// the directory is: the engine's commands go to the data's node.
func TestStorageOnDataNode(t *testing.T) {
	shared := &backup.Storage{
		RepositoryPassword: "pw",
		StorageLocal:       &backup.StorageLocal{Path: "/host/nfs", NodeID: "node-1", Shared: true},
	}
	moved := storageOnDataNode(shared, "", "zone=b")
	assert.Equal(t, &backup.StorageLocal{Path: "/host/nfs", NodeLabel: "zone=b", Shared: true}, moved.StorageLocal)
	assert.Equal(t, "pw", moved.RepositoryPassword)
	assert.Equal(t, "node-1", shared.StorageLocal.NodeID, "the repository's own storage is left alone")

	pinned := &backup.Storage{StorageLocal: &backup.StorageLocal{Path: "/host/srv", NodeID: "node-2"}}
	assert.Same(t, pinned, storageOnDataNode(pinned, "node-2", ""))
	cloud := &backup.Storage{StorageS3: &backup.StorageS3{Bucket: "b"}}
	assert.Same(t, cloud, storageOnDataNode(cloud, "node-3", ""))
}
