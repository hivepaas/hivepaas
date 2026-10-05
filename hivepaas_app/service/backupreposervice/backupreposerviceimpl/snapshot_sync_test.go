package backupreposerviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

// A snapshot's setting is created when the snapshot was taken - one read from
// a repository long after, too: the list is ordered by it. Without a time, when
// it is recorded.
func TestASnapshotSettingIsCreatedWhenTheSnapshotWasTaken(t *testing.T) {
	taken := time.Date(2026, 8, 1, 2, 3, 4, 500_000_000, time.UTC)
	before := time.Now()
	settings, _ := NewSnapshotSettings(&entity.Setting{ID: "repo"}, []*backupreposervice.RepoSnapshot{
		{Snapshot: &entity.BackupSnapshot{ID: "s1", ShortID: "s1", Time: taken}},
		{Snapshot: &entity.BackupSnapshot{ID: "s2", ShortID: "s2"}},
	})

	if assert.Len(t, settings, 2) {
		assert.Equal(t, taken, settings[0].CreatedAt)
		assert.False(t, settings[0].UpdatedAt.Before(before), "recorded now")
		assert.False(t, settings[1].CreatedAt.Before(before), "no time: when recorded")
	}
}
