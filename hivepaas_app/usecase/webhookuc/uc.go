package webhookuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appRepo        repository.AppRepo
	deploymentRepo repository.DeploymentRepo
	settingRepo    repository.SettingRepo

	appDeploymentService appdeploymentservice.Service
	appPreviewService    apppreviewservice.Service
	appService           appservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	deploymentRepo repository.DeploymentRepo,
	settingRepo repository.SettingRepo,

	appDeploymentService appdeploymentservice.Service,
	appPreviewService apppreviewservice.Service,
	appService appservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		appRepo:        appRepo,
		deploymentRepo: deploymentRepo,
		settingRepo:    settingRepo,

		appDeploymentService: appDeploymentService,
		appPreviewService:    appPreviewService,
		appService:           appService,
	}
}
