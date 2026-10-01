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

// WithheldSecret is a secret of an app its previews go without, not being
// inheritable, and the app's variables that use it: in a preview they are empty.
type WithheldSecret struct {
	Name    string
	EnvVars []string
}
