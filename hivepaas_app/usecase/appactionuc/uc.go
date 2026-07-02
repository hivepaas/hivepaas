package appactionuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
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
