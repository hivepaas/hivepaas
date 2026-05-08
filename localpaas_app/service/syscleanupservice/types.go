package syscleanupservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type SysCleanupReq struct {
	*queue.TaskExecData
	CronJob *entity.Setting
}

type SysCleanupResp struct {
	SkipResultNotification bool
}
