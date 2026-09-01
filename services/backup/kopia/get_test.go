package kopia

import (
	"context"
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
