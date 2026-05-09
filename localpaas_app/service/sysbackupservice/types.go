package sysbackupservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type SysBackupReq struct {
	*queue.TaskExecData
	SysBackupSettings *entity.SystemBackup
}

type SysBackupResp struct {
	SkipResultNotification bool
}
