package projectserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	db *database.DB,

	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	fileRepo repository.FileRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	tagRepo repository.TagRepo,
	taskRepo repository.TaskRepo,
	userRepo repository.UserRepo,

	appService appservice.Service,
	envVarService envvarservice.Service,
	networkService networkservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) projectservice.Service {
	return &service{
		db: db,

		appRepo:        appRepo,
		binObjectRepo:  binObjectRepo,
		fileRepo:       fileRepo,
		projectEnvRepo: projectEnvRepo,
		projectRepo:    projectRepo,
		resLinkRepo:    resLinkRepo,
		settingRepo:    settingRepo,
		tagRepo:        tagRepo,
		taskRepo:       taskRepo,
		userRepo:       userRepo,

		appService:     appService,
		envVarService:  envVarService,
		networkService: networkService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}

type service struct {
	db *database.DB

	appRepo        repository.AppRepo
	binObjectRepo  repository.BinObjectRepo
	fileRepo       repository.FileRepo
	projectEnvRepo repository.ProjectEnvRepo
	projectRepo    repository.ProjectRepo
	resLinkRepo    repository.ResLinkRepo
	settingRepo    repository.SettingRepo
	tagRepo        repository.TagRepo
	taskRepo       repository.TaskRepo
	userRepo       repository.UserRepo

	appService     appservice.Service
	envVarService  envvarservice.Service
	networkService networkservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}
