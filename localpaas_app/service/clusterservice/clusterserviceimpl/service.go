package clusterserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	dockerManager docker.Manager,
) clusterservice.Service {
	return &service{
		appRepo:     appRepo,
		settingRepo: settingRepo,

		dockerManager: dockerManager,
	}
}

type service struct {
	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	dockerManager docker.Manager
}
