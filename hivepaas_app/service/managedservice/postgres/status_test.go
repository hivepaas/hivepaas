package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

var statusNow = time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)

func statusManaged(mode string) *entity.AppManagedSettings {
	return &entity.AppManagedSettings{Engine: "postgres", EngineAPI: 1, Postgres: &entity.ManagedPostgres{
		Storage:     entity.ManagedStorage{Placement: "same-node"},
		RestoreTest: entity.ManagedRestoreTest{Mode: mode},
		Jobs: entity.ManagedPostgresJobs{Nightly: entity.ObjectID{ID: "nightly"},
			Health: entity.ObjectID{ID: "health"}, RestoreTest: entity.ObjectID{ID: "test"}},
	}}
}

func healthyRuns() map[string]*managedmodel.JobRun {
	return map[string]*managedmodel.JobRun{
		"health": {Status: base.TaskStatusDone, EndedAt: statusNow.Add(-2 * time.Minute), Outputs: map[string]string{
			outLastArchivedAt: "2026-10-18T11:59:30Z", outFailedCount: "0", outPgWalBytes: "33554432",
			outDBBytes: "1073741824", outFreeBytes: "10737418240",
		}},
		"nightly": {Status: base.TaskStatusDone, EndedAt: statusNow.Add(-10 * time.Hour), Outputs: map[string]string{
			outBackupName: "base_000000010000000000000042", outBackupAt: "2026-10-18T02:00:05Z",
			outBackupBytes: "52428800", outOldestBackupAt: "2026-10-11T02:00:04Z",
		}},
		"test": {Status: base.TaskStatusDone, EndedAt: statusNow.Add(-32 * time.Hour), Outputs: map[string]string{
			outRestoredTo: "hp-test-20261017T0400", outDurationSeconds: "95",
		}},
	}
}

func TestStatusOfAHealthyDatabase(t *testing.T) {
	status, err := New().Status(&managedmodel.StatusReq{Managed: statusManaged(entity.ManagedRestoreTestContainer),
		Runs: healthyRuns(), Now: statusNow})

	assert.NoError(t, err)
	assert.Equal(t, managedmodel.ArchivingOK, status.Archiving.State)
	assert.Equal(t, time.Date(2026, 10, 18, 11, 59, 30, 0, time.UTC), *status.Archiving.LastArchivedAt)
	assert.Equal(t, int64(33554432), status.Archiving.PgWalBytes)
	assert.False(t, status.Archiving.Stale)
	if assert.NotNil(t, status.RestoreWindow) {
		assert.Equal(t, time.Date(2026, 10, 11, 2, 0, 4, 0, time.UTC), status.RestoreWindow.From)
		assert.Equal(t, time.Date(2026, 10, 18, 11, 59, 30, 0, time.UTC), status.RestoreWindow.To)
	}
	if assert.NotNil(t, status.LastBackup) {
		assert.True(t, status.LastBackup.Succeeded)
		assert.Equal(t, "base_000000010000000000000042", status.LastBackup.Name)
		assert.Equal(t, int64(52428800), status.LastBackup.Bytes)
	}
	if assert.NotNil(t, status.LastRestoreTest) {
		assert.True(t, status.LastRestoreTest.Passed)
		assert.Equal(t, 95*time.Second, status.LastRestoreTest.Duration)
		assert.Equal(t, "hp-test-20261017T0400", status.LastRestoreTest.RestoredTo)
	}
	assert.Equal(t, "same-node", status.Placement)
	assert.Empty(t, status.Suggestions)
}

func TestStatusOfAFailingArchive(t *testing.T) {
	runs := healthyRuns()
	runs["health"].Status = base.TaskStatusFailed
	runs["health"].Outputs[outFailedCount] = "12"

	status, err := New().Status(&managedmodel.StatusReq{Managed: statusManaged(entity.ManagedRestoreTestContainer),
		Runs: runs, Now: statusNow})

	assert.NoError(t, err)
	assert.Equal(t, managedmodel.ArchivingFailing, status.Archiving.State)
	assert.Equal(t, 12, status.Archiving.FailedCount)
}

// A health check that has not run for a while says nothing about now.
func TestStatusOfAHealthCheckThatStopped(t *testing.T) {
	runs := healthyRuns()
	runs["health"].EndedAt = statusNow.Add(-time.Hour)

	status, _ := New().Status(&managedmodel.StatusReq{Managed: statusManaged(entity.ManagedRestoreTestContainer),
		Runs: runs, Now: statusNow})

	assert.True(t, status.Archiving.Stale)
}

func TestStatusOfANewDatabase(t *testing.T) {
	status, err := New().Status(&managedmodel.StatusReq{Managed: statusManaged(entity.ManagedRestoreTestContainer),
		Now: statusNow})

	assert.NoError(t, err)
	assert.Equal(t, managedmodel.ArchivingUnknown, status.Archiving.State)
	assert.Nil(t, status.RestoreWindow, "no base backup yet: nothing to restore from")
	assert.Nil(t, status.LastBackup)
	assert.Nil(t, status.LastRestoreTest)
}

// The verifier is suggested when the container's test would not fit, or keeps
// the database busy too long - not before, since a new database has no size.
func TestStatusSuggestsTheVerifier(t *testing.T) {
	for name, change := range map[string]func(runs map[string]*managedmodel.JobRun){
		"no room": func(runs map[string]*managedmodel.JobRun) {
			runs["health"].Outputs[outFreeBytes] = "1073741824" // 1 GiB free for a 1 GiB database
		},
		"too long": func(runs map[string]*managedmodel.JobRun) {
			runs["test"].Outputs[outDurationSeconds] = "2400"
		},
	} {
		runs := healthyRuns()
		change(runs)

		status, _ := New().Status(&managedmodel.StatusReq{Managed: statusManaged(entity.ManagedRestoreTestContainer),
			Runs: runs, Now: statusNow})
		assert.Equal(t, []string{managedmodel.SuggestVerifier}, status.Suggestions, name)

		status, _ = New().Status(&managedmodel.StatusReq{Managed: statusManaged(entity.ManagedRestoreTestVerifier),
			Runs: runs, Now: statusNow})
		assert.Empty(t, status.Suggestions, "%s: it already has one", name)
	}
}
