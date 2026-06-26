package clusterserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	permissionManager permission.Manager,
	dockerManager docker.Manager,
) clusterservice.Service {
	return &service{
		settingRepo:       settingRepo,
		appRepo:           appRepo,
		permissionManager: permissionManager,
		dockerManager:     dockerManager,
	}
}

type service struct {
	settingRepo       repository.SettingRepo
	appRepo           repository.AppRepo
	permissionManager permission.Manager
	dockerManager     docker.Manager
}
