package backuprepocleanupservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type BackupRepoCleanupReq struct {
	*queue.TaskExecData
	CleanupJobSetting *entity.Setting
	CleanupSettings   *entity.BackupRepoCleanup
}

type BackupRepoCleanupResp struct {
	TaskOutput             *entity.TaskBackupRepoCleanupOutput
	SkipResultNotification bool
}
