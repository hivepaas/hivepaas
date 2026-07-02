package sysbackupservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SysBackupReq struct {
	*queue.TaskExecData
	SysBackupSettings *entity.SystemBackup
}

type SysBackupResp struct {
	SkipResultNotification bool
}
