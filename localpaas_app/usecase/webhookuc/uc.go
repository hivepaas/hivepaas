package webhookuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/apppreviewservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	deploymentRepo repository.DeploymentRepo
	settingRepo    repository.SettingRepo

	appDeploymentService appdeploymentservice.Service
	appPreviewService    apppreviewservice.Service
	appService           appservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	deploymentRepo repository.DeploymentRepo,
	settingRepo repository.SettingRepo,

	appDeploymentService appdeploymentservice.Service,
	appPreviewService apppreviewservice.Service,
	appService appservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		deploymentRepo: deploymentRepo,
		settingRepo:    settingRepo,

		appDeploymentService: appDeploymentService,
		appPreviewService:    appPreviewService,
		appService:           appService,
	}
}
