package taskdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A restore's task says what it restored, from where, and how: its page links
// to the snapshot.
func TestTransformTaskGivesARestoresSnapshot(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeBackupRestore, TargetID: "s1"}
	task.MustSetArgs(&entity.TaskBackupRestoreArgs{
		RepoID: "r1", SnapshotID: "k1full", SnapshotPath: "uploads", Mode: base.BackupRestoreModeReplace,
		StopApp: true,
	})

	resp, err := TransformTask(task, nil, entity.NewRefObjects())

	assert.NoError(t, err)
	assert.Equal(t, &TaskBackupRestoreResp{
		SnapshotRecordID: "s1", RepoID: "r1", SnapshotID: "k1full", SnapshotPath: "uploads",
		Mode: base.BackupRestoreModeReplace, StopApp: true,
	}, resp.BackupRestore)
	assert.Nil(t, resp.DataBackup)
}
