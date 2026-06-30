package appactionuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appDeploymentService appdeploymentservice.Service
	appService           appservice.Service
	clusterService       clusterservice.Service
	settingService       settingservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appDeploymentService appdeploymentservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	settingService settingservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		appDeploymentService: appDeploymentService,
		appService:           appService,
		clusterService:       clusterService,
		settingService:       settingService,

		dockerManager: dockerManager,
	}
}
