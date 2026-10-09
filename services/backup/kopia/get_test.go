package kopia

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func TestClient_ListSnapshots_Filters(t *testing.T) {
	mockJSON := `[
		{
			"id": "snap1",
			"source": {"host": "node-1", "path": "/var/lib/data1"},
			"startTime": "2026-08-28T10:00:00Z",
			"tags": {"app": "app-1", "env": "prod"},
			"stats": {"totalSize": 100}
		},
		{
			"id": "snap2",
			"source": {"host": "node-2", "path": "/var/lib/data2"},
			"startTime": "2026-08-28T12:00:00Z",
			"tags": {"app": "app-2", "env": "staging"},
			"stats": {"totalSize": 200}
		}
	]`

	mockExecutor := func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		if req.Stdout != nil {
			_, _ = req.Stdout.Write([]byte(mockJSON))
		}
		return &backupmodel.CommandExecResp{ExitCode: 0}, nil
	}

	c := NewClient(&backupmodel.Storage{
		RepositoryPassword: "secret-password",
		StorageLocal: &backupmodel.StorageLocal{
			Path: "/mnt/backups",
		},
	}, mockExecutor)

	// Filter by Hostname
	res, err := c.ListSnapshots(context.Background(), &backupmodel.ListSnapshotsOptions{
		Hostname: "node-1",
	})
	assert.NoError(t, err)
	assert.Len(t, res.Items, 1)
	assert.Equal(t, "snap1", res.Items[0].ID)

	// Filter by Path
	res, err = c.ListSnapshots(context.Background(), &backupmodel.ListSnapshotsOptions{
		Path: "/var/lib/data2",
	})
	assert.NoError(t, err)
	assert.Len(t, res.Items, 1)
	assert.Equal(t, "snap2", res.Items[0].ID)

	// Filter by Tags
	res, err = c.ListSnapshots(context.Background(), &backupmodel.ListSnapshotsOptions{
		Tags: []string{"app:app-1"},
	})
	assert.NoError(t, err)
	assert.Len(t, res.Items, 1)
	assert.Equal(t, "snap1", res.Items[0].ID)

	// Filter by Limit
	res, err = c.ListSnapshots(context.Background(), &backupmodel.ListSnapshotsOptions{
		Limit: 1,
	})
	assert.NoError(t, err)
	assert.Len(t, res.Items, 1)
}

func TestToStandardSnapshot(t *testing.T) {
	now := time.Now()
	manifest := &kopiaSnapshotManifest{
		ID: "k1234567890abcdef",
		Source: kopiaSource{
			Host: "worker-node-1",
			Path: "/var/lib/docker/volumes/app_data",
		},
		StartTime: now,
		Tags: map[string]string{
			"app": "my-app",
			"env": "production",
		},
		Stats: kopiaStats{
			TotalSize:      20971520,
			TotalFileCount: 42,
		},
	}

	snap := toStandardSnapshot(manifest)
	assert.Equal(t, "k1234567890abcdef", snap.ID)
	assert.Equal(t, "k1234567", snap.ShortID)
	assert.Equal(t, now, snap.Time)
	assert.Equal(t, "worker-node-1", snap.Hostname)
	assert.Equal(t, int64(20971520), snap.SizeBytes)
	assert.Contains(t, snap.Tags, "app:my-app")
	assert.Contains(t, snap.Tags, "env:production")
	assert.Contains(t, snap.Paths, "/var/lib/docker/volumes/app_data")
}

// Kopia writes user tags into the manifest under a "tag:" namespace. Storing that prefix would
// mean a tag search for "app:my-app" never matches what was actually saved.
func TestToStandardSnapshot_StripsKopiaTagNamespace(t *testing.T) {
	snap := toStandardSnapshot(&kopiaSnapshotManifest{
		ID:        "k1234567890abcdef",
		StartTime: time.Now(),
		Tags: map[string]string{
			"tag:app": "my-app",
			"tag:env": "production",
			"plain":   "",
		},
	})

	assert.Equal(t, []string{"app:my-app", "env:production", "plain"}, snap.Tags)
}

func TestHasMatchingTag(t *testing.T) {
	snapshotTags := []string{"app:web", "env:prod", "type:db"}

	assert.True(t, hasMatchingTag(snapshotTags, []string{"env:prod"}))
	assert.True(t, hasMatchingTag(snapshotTags, []string{"other", "type:db"}))
	assert.False(t, hasMatchingTag(snapshotTags, []string{"env:staging"}))
}

func TestHasMatchingPath(t *testing.T) {
	paths := []string{"/var/lib/docker/volumes/my_data/"}

	assert.True(t, hasMatchingPath(paths, "/var/lib/docker/volumes/my_data"))
	assert.True(t, hasMatchingPath(paths, "/var/lib/docker/volumes/my_data/"))
	assert.False(t, hasMatchingPath(paths, "/var/lib/docker/volumes/other"))
}

// `snapshot create --json` gives no stats: the size is the root entry's sum.
func TestToStandardSnapshot_SizeFromTheRootEntry(t *testing.T) {
	manifest := &kopiaSnapshotManifest{ID: "k1", Description: "nightly (run t1)"}
	manifest.RootEntry.Summary.Size = 42

	snap := toStandardSnapshot(manifest)
	assert.Equal(t, int64(42), snap.SizeBytes)
	assert.Equal(t, "nightly (run t1)", snap.Description)
}

// A snapshot is read by its manifest - `snapshot list` finds none by its ID -
// and its root object comes with it.
func TestClient_GetSnapshot_ReadsTheManifest(t *testing.T) {
	var args []string
	answer := "// id: 7ef3a908\n// length: 732\n// label type:snapshot\n// label username:hivepaas\n" +
		`{"id":"","source":{"host":"data-backup","userName":"hivepaas","path":"/j2"},` +
		`"startTime":"2026-10-09T10:00:00Z","rootEntry":{"obj":"kroot1","summ":{"size":42}}}` + "\n"
	c := NewClient(&backupmodel.Storage{StorageLocal: &backupmodel.StorageLocal{Path: "/mnt/backups"}},
		func(_ context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
			args = req.Command
			_, _ = req.Stdout.Write([]byte(answer))
			return &backupmodel.CommandExecResp{}, nil
		})

	got, err := c.GetSnapshot(context.Background(), "7ef3a908")

	assert.NoError(t, err)
	if assert.NotNil(t, got.Item) {
		assert.Equal(t, "7ef3a908", got.Item.ID)
		assert.Equal(t, "kroot1", got.Item.RootObjectID)
		assert.Equal(t, int64(42), got.Item.SizeBytes)
		assert.Equal(t, []string{"/j2"}, got.Item.Paths)
	}
	assert.Contains(t, strings.Join(args, " "), "manifest show 7ef3a908")
}

// A manifest that is not a snapshot's is no snapshot.
func TestClient_GetSnapshot_OfAnotherManifest(t *testing.T) {
	c := NewClient(&backupmodel.Storage{StorageLocal: &backupmodel.StorageLocal{Path: "/mnt/backups"}},
		func(_ context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
			_, _ = req.Stdout.Write([]byte("// id: p1\n// label type:policy\n{\"actions\":{}}\n"))
			return &backupmodel.CommandExecResp{}, nil
		})

	_, err := c.GetSnapshot(context.Background(), "p1")

	assert.ErrorIs(t, err, backupmodel.ErrSnapshotNotFound)
}
