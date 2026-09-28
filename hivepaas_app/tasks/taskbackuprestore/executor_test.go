package taskbackuprestore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// A restore's arguments say which of the two it is: a command's, or a volume's.
func TestRestoreReqOf(t *testing.T) {
	exec := &queue.TaskExecData{Task: &entity.Task{ID: "t1"}}
	app := &entity.App{ID: "app1"}
	target := backupreposervice.RepoTarget{RepoSetting: &entity.Setting{ID: "repo1"}}

	command := &entity.TaskBackupRestoreArgs{
		SnapshotID: "k1", Command: &entity.CommandTemplate{Command: "psql app"}, FileName: "db.sql",
	}
	assert.Equal(t, &databackupservice.RestoreReq{
		TaskExecData: exec, Target: target, SnapshotID: "k1", App: app,
		Command: &databackupservice.RestoreCommand{Command: command.Command, FileName: "db.sql"},
	}, restoreReqOf(exec, command, app, target))

	volume := &entity.TaskBackupRestoreArgs{
		SnapshotID: "k1", Volume: entity.ObjectID{ID: "vol1"}, Subpath: "data", SnapshotPath: "uploads",
		StopApp: true, Mode: base.BackupRestoreModeReplace,
	}
	assert.Equal(t, &databackupservice.RestoreReq{
		TaskExecData: exec, Target: target, SnapshotID: "k1", App: app,
		Volume: &databackupservice.RestoreVolume{
			VolumeID: "vol1", Subpath: "data", SnapshotPath: "uploads", StopApp: true,
			Mode: base.BackupRestoreModeReplace,
		},
	}, restoreReqOf(exec, volume, app, target))
}
