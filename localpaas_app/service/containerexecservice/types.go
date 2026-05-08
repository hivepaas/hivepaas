package containerexecservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type ContainerExecReq struct {
	*queue.TaskExecData
	CronJob *entity.Setting
	Project *entity.Project
	App     *entity.App
}

type ContainerExecResp struct {
	SkipResultNotification bool
}
