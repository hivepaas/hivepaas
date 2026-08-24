package apppreviewservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type CreatePreviewReq struct {
	*queue.TaskExecData

	App              *entity.App
	OnInitDeployment func(*entity.Deployment) error
	OnDeploymentTask func(*entity.Task) error
}

type CreatePreviewResp struct {
	PreviewApp     *entity.App
	Deployment     *entity.Deployment
	DeploymentTask *entity.Task
	OnCleanup      func(error) error
}
