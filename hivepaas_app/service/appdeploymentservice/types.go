package appdeploymentservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type AppDeploymentReq struct {
	*queue.TaskExecData
}

type AppDeploymentResp struct {
	Deployment *entity.Deployment
}
