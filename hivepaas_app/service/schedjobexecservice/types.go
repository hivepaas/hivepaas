package schedjobexecservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SchedJobExecReq struct {
	*queue.TaskExecData
	SchedJobSetting        *entity.Setting
	Project                *entity.Project
	App                    *entity.App
	TaskMinRunningDuration time.Duration
	TaskFindRetryMax       int
	TaskFindRetryDelay     time.Duration
}

type SchedJobExecResp struct {
	SkipResultNotification bool
}
