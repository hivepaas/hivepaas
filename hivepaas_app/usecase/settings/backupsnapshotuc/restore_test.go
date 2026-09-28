package backupsnapshotuc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
	"github.com/hivepaas/hivepaas/services/backup"
)

func rootOf(entries ...backup.SnapshotEntry) func() ([]backup.SnapshotEntry, error) {
	return func() ([]backup.SnapshotEntry, error) { return entries, nil }
}

func notListed(t *testing.T) func() ([]backup.SnapshotEntry, error) {
	return func() ([]backup.SnapshotEntry, error) {
		t.Error("the snapshot was listed, though its tag and job say what it is")
		return nil, nil
	}
}

// A snapshot's kind is its tag's or its job's; one taken outside HivePaaS goes
// by what it holds: one file is a command's.
func TestSnapshotKind(t *testing.T) {
	kind, err := snapshotKind(base.SchedJobDataBackupSourceCommand, "db.sql", notListed(t))
	assert.NoError(t, err)
	assert.Equal(t, &restoreKind{command: true, fileName: "db.sql"}, kind)

	kind, err = snapshotKind(base.SchedJobDataBackupSourceVolume, "", notListed(t))
	assert.NoError(t, err)
	assert.Equal(t, &restoreKind{}, kind)

	// Its job is gone: the file is the one it holds.
	kind, err = snapshotKind(base.SchedJobDataBackupSourceCommand, "", rootOf(backup.SnapshotEntry{Name: "dump.sql"}))
	assert.NoError(t, err)
	assert.Equal(t, &restoreKind{command: true, fileName: "dump.sql"}, kind)

	kind, err = snapshotKind("", "", rootOf(backup.SnapshotEntry{Name: "db.sql", SizeBytes: 3}))
	assert.NoError(t, err)
	assert.Equal(t, &restoreKind{command: true, fileName: "db.sql"}, kind)

	kind, err = snapshotKind("", "", rootOf(backup.SnapshotEntry{Name: "uploads", Dir: true}))
	assert.NoError(t, err)
	assert.Equal(t, &restoreKind{}, kind)

	kind, err = snapshotKind("", "", rootOf(backup.SnapshotEntry{Name: "a"}, backup.SnapshotEntry{Name: "b"}))
	assert.NoError(t, err)
	assert.Equal(t, &restoreKind{}, kind)

	_, err = snapshotKind(base.SchedJobDataBackupSourceCommand, "", rootOf(backup.SnapshotEntry{Name: "d", Dir: true}))
	assert.Error(t, err, "a command's snapshot holds one file")

	_, err = snapshotKind("", "", func() ([]backup.SnapshotEntry, error) { return nil, errors.New("gone") })
	assert.ErrorContains(t, err, "gone")
}

// A restore is the snapshot's kind: a command for a command's, a volume for a
// volume's.
func TestRestoreFitsTheSnapshot(t *testing.T) {
	command := &backupsnapshotdto.RestoreBackupSnapshotReq{
		Command: &commandtemplatedto.CommandTemplateBaseReq{Command: "psql"},
	}
	volume := &backupsnapshotdto.RestoreBackupSnapshotReq{Volume: basedto.ObjectIDReq{ID: "vol1"}}

	assert.NoError(t, restoreFits(&restoreKind{command: true}, command))
	assert.NoError(t, restoreFits(&restoreKind{}, volume))
	assert.ErrorIs(t, restoreFits(&restoreKind{}, command), hperrors.ErrArgumentInvalid)
	assert.ErrorIs(t, restoreFits(&restoreKind{command: true}, volume), hperrors.ErrArgumentInvalid)
}

// An app has one restore at a time: those not ended are looked for.
func TestRestoreInFlightQuery(t *testing.T) {
	var tasks []*entity.Task
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&tasks),
		restoreInFlightOpts("app1")...).String()

	assert.Contains(t, sql, `task.type = 'task:backup-restore'`)
	assert.Contains(t, sql, `task.object_id = 'app1'`)
	assert.Contains(t, sql, `task.status IN ('not-started', 'in-progress')`)
}

// The task names the snapshot's record as its target and the app as its
// object, and carries the restore as it was asked for.
func TestRestoreTask(t *testing.T) {
	now := time.Date(2026, 9, 28, 15, 30, 0, 0, time.UTC)
	app := &entity.App{ID: "app1", ProjectID: "p1"}
	args := &entity.TaskBackupRestoreArgs{ProjectID: "p1", AppID: "app1", RepoID: "r1", SnapshotID: "k1full"}

	task := restoreTask("s1", app, args, now)

	assert.Equal(t, base.ObjectScopeApp, task.Scope)
	assert.Equal(t, "app1", task.ObjectID)
	assert.Equal(t, "s1", task.TargetID)
	assert.Equal(t, base.TaskTypeBackupRestore, task.Type)
	assert.Equal(t, base.TaskStatusNotStarted, task.Status)
	assert.Equal(t, now, task.RunAt)
	got, err := task.ArgsAsBackupRestore()
	assert.NoError(t, err)
	assert.Equal(t, args, got)
}
