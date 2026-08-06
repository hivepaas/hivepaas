package commandpipeexecservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type CommandPipeExecReq struct {
	*queue.TaskExecData
	CommandPipes           []*entity.Setting
	SrcApp                 *entity.App
	DestApp                *entity.App
	TaskMinRunningDuration time.Duration
	TaskFindRetryMax       int
	TaskFindRetryDelay     time.Duration
}

type CommandPipeExecResp struct {
}
