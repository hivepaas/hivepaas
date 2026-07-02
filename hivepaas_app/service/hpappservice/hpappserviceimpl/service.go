package hpappserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	taskRepo repository.TaskRepo,

	dockerManager docker.Manager,
) hpappservice.Service {
	return &service{
		taskRepo: taskRepo,

		dockerManager: dockerManager,
	}
}

type service struct {
	taskRepo repository.TaskRepo

	dockerManager docker.Manager
}
