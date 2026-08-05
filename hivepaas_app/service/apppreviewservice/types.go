package apppreviewservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type CreatePreviewReq struct {
	App             *entity.App
	RepoRef         string
	CustomSubdomain string
	NoStart         bool
	CloneDBApps     []*entity.App

	OnInitDeployment func(*entity.Deployment) error
	OnDeploymentTask func(*entity.Task) error
}

type CreatePreviewResp struct {
	PreviewApp     *entity.App
	Deployment     *entity.Deployment
	DeploymentTask *entity.Task
	OnCleanup      func(error) error
}
