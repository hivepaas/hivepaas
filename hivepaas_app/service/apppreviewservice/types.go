package apppreviewservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type CreatePreviewReq struct {
	App             *entity.App
	RepoRef         string
	CustomSubdomain string
	NoStart         bool
	CloneDBApps     []*entity.App

	OnInitDeployment func(*entity.Deployment) error
	OnDeploymentTask func(*entity.Task) error

	RefObjects *entity.RefObjects
	LogStore   *tasklog.Store
}

type CreatePreviewResp struct {
	PreviewApp     *entity.App
	Deployment     *entity.Deployment
	DeploymentTask *entity.Task
	OnCleanup      func(error) error
}
