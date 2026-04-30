package lpappserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	dockerManager docker.Manager,
) lpappservice.Service {
	return &service{
		settingRepo:   settingRepo,
		taskRepo:      taskRepo,
		dockerManager: dockerManager,
	}
}

type service struct {
	settingRepo   repository.SettingRepo
	taskRepo      repository.TaskRepo
	dockerManager docker.Manager
}
