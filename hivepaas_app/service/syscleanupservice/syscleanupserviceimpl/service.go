package syscleanupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	deploymentRepo repository.DeploymentRepo
	fileRepo       repository.FileRepo
	lockRepo       repository.LockRepo
	sysErrorRepo   repository.SysErrorRepo
	taskLogRepo    repository.TaskLogRepo
	taskRepo       repository.TaskRepo

	dockerManager docker.Manager
}

func New(
	deploymentRepo repository.DeploymentRepo,
	fileRepo repository.FileRepo,
	lockRepo repository.LockRepo,
	sysErrorRepo repository.SysErrorRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,

	dockerManager docker.Manager,
) syscleanupservice.Service {
	return &service{
		deploymentRepo: deploymentRepo,
		fileRepo:       fileRepo,
		lockRepo:       lockRepo,
		sysErrorRepo:   sysErrorRepo,
		taskLogRepo:    taskLogRepo,
		taskRepo:       taskRepo,

		dockerManager: dockerManager,
	}
}
