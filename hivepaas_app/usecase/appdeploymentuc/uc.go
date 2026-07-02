package appdeploymentuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

type UC struct {
	db *database.DB

	deploymentInfoRepo cacherepository.DeploymentInfoRepo
	deploymentRepo     repository.DeploymentRepo
	taskControlRepo    cacherepository.TaskControlRepo

	taskService taskservice.Service
	userService userservice.Service
}

func New(
	db *database.DB,

	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	deploymentRepo repository.DeploymentRepo,
	taskControlRepo cacherepository.TaskControlRepo,

	taskService taskservice.Service,
	userService userservice.Service,
) *UC {
	return &UC{
		db: db,

		deploymentInfoRepo: deploymentInfoRepo,
		deploymentRepo:     deploymentRepo,
		taskControlRepo:    taskControlRepo,

		taskService: taskService,
		userService: userService,
	}
}
