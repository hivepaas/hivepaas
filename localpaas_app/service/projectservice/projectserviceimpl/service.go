package projectserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	fileRepo repository.FileRepo,
	projectRepo repository.ProjectRepo,
	projectTagRepo repository.ProjectTagRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	userRepo repository.UserRepo,

	appService appservice.Service,
	networkService networkservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) projectservice.Service {
	return &service{
		appRepo:        appRepo,
		binObjectRepo:  binObjectRepo,
		fileRepo:       fileRepo,
		projectRepo:    projectRepo,
		projectTagRepo: projectTagRepo,
		resLinkRepo:    resLinkRepo,
		settingRepo:    settingRepo,
		taskRepo:       taskRepo,
		userRepo:       userRepo,

		appService:     appService,
		networkService: networkService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}

type service struct {
	appRepo        repository.AppRepo
	binObjectRepo  repository.BinObjectRepo
	fileRepo       repository.FileRepo
	projectRepo    repository.ProjectRepo
	projectTagRepo repository.ProjectTagRepo
	resLinkRepo    repository.ResLinkRepo
	settingRepo    repository.SettingRepo
	taskRepo       repository.TaskRepo
	userRepo       repository.UserRepo

	appService     appservice.Service
	networkService networkservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}
