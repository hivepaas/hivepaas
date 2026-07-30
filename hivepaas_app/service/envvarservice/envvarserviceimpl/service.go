package envvarserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	clusterService clusterservice.Service,
	projectService projectservice.Service,
	dockerManager docker.Manager,
) envvarservice.Service {
	return &service{
		appRepo:     appRepo,
		resLinkRepo: resLinkRepo,
		settingRepo: settingRepo,

		appService:     appService,
		clusterService: clusterService,
		projectService: projectService,
		dockerManager:  dockerManager,
	}
}

type service struct {
	appRepo     repository.AppRepo
	resLinkRepo repository.ResLinkRepo
	settingRepo repository.SettingRepo

	appService     appservice.Service
	clusterService clusterservice.Service
	projectService projectservice.Service
	dockerManager  docker.Manager
}
