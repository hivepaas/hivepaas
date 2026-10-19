package mariadb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

var statusNow = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func statusManaged() *entity.AppManagedSettings {
	return &entity.AppManagedSettings{Engine: "mariadb", EngineAPI: 1, Database: &entity.ManagedDatabase{
		RetainFull: 2,
		Jobs: entity.ManagedJobs{Nightly: entity.ObjectID{ID: "nightly"},
			Health: entity.ObjectID{ID: "health"}, RestoreTest: entity.ObjectID{ID: "test"}},
	}}
}

// The status is read from the outputs MariaDB's scripts write: its health
// check's pending binary logs, and its nightly run's stream backups.
func TestStatusReadsMariaDBsOutputs(t *testing.T) {
	runs := map[string]*managedmodel.JobRun{
		"health": {Status: base.TaskStatusDone, EndedAt: statusNow.Add(-time.Minute), Outputs: map[string]string{
			"LAST_ARCHIVED_AT": "2026-10-19T11:59:00Z", "FAILED_COUNT": "0", "PENDING_BYTES": "2264",
			"DB_BYTES": "5931008", "FREE_BYTES": "52120552652", "READY_COUNT": "1", "REPLICAS": "",
		}},
		"nightly": {Status: base.TaskStatusDone, EndedAt: statusNow.Add(-10 * time.Hour), Outputs: map[string]string{
			"BACKUP_NAME": "stream_20261019T020001Z", "BACKUP_AT": "2026-10-19T02:00:01Z", "BACKUP_BYTES": "1636449",
			"OLDEST_BACKUP_AT": "2026-10-18T02:00:01Z",
			"BACKUPS": "stream_20261018T020001Z@2026-10-18T02:00:01Z@1636000 " +
				"stream_20261019T020001Z@2026-10-19T02:00:01Z@1636449 not-a-backup@2026-10-19T02:00:01Z@1",
		}},
	}

	status, err := New().Status(&managedmodel.StatusReq{Managed: statusManaged(), Runs: runs, Now: statusNow})

	assert.NoError(t, err)
	assert.Equal(t, managedmodel.ArchivingOK, status.Archiving.State)
	assert.Equal(t, int64(2264), status.Archiving.PendingBytes)
	assert.Equal(t, time.Date(2026, 10, 19, 11, 59, 0, 0, time.UTC), *status.Archiving.LastArchivedAt)
	assert.Equal(t, "stream_20261019T020001Z", status.LastBackup.Name)
	if assert.Len(t, status.Backups, 2, "what is not a backup's name is left out") {
		assert.Equal(t, "stream_20261018T020001Z", status.Backups[0].Name)
	}
	if assert.NotNil(t, status.RestoreWindow) {
		assert.Equal(t, time.Date(2026, 10, 18, 2, 0, 1, 0, time.UTC), status.RestoreWindow.From)
		assert.Equal(t, time.Date(2026, 10, 19, 11, 59, 0, 0, time.UTC), status.RestoreWindow.To)
	}
	assert.Nil(t, status.Members, "no replicas")
}

// MariaDB's backups are full ones, named by time: retention keeps the newest
// retainFull, and deletes before the oldest of them.
func TestRetentionKeepsTheNewestStreamBackups(t *testing.T) {
	outputs := map[string]string{"BACKUPS": "stream_20261017T020001Z@2026-10-17T02:00:01Z@1 " +
		"stream_20261018T020001Z@2026-10-18T02:00:01Z@1 stream_20261019T020001Z@2026-10-19T02:00:01Z@1"}

	retention, err := New().Retention(&managedmodel.RetentionReq{Managed: statusManaged(), Outputs: outputs,
		Now: statusNow})

	assert.NoError(t, err)
	assert.Equal(t, "retain stream_20261018T020001Z", retention.Args)
	assert.Equal(t, "stream_20261018T020001Z", retention.Boundary)
	assert.Equal(t, "backup", New().NightlyBackupArgs())
}
