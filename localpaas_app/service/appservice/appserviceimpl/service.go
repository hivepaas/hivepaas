package appserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	db *database.DB,

	appRepo repository.AppRepo,
	appTagRepo repository.AppTagRepo,
	deploymentRepo repository.DeploymentRepo,
	fileRepo repository.FileRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	clusterService clusterservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) appservice.Service {
	return &service{
		db: db,

		appRepo:        appRepo,
		appTagRepo:     appTagRepo,
		deploymentRepo: deploymentRepo,
		fileRepo:       fileRepo,
		resLinkRepo:    resLinkRepo,
		settingRepo:    settingRepo,
		taskRepo:       taskRepo,

		clusterService: clusterService,
		traefikService: traefikService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}

type service struct {
	db *database.DB

	appRepo        repository.AppRepo
	appTagRepo     repository.AppTagRepo
	deploymentRepo repository.DeploymentRepo
	fileRepo       repository.FileRepo
	resLinkRepo    repository.ResLinkRepo
	settingRepo    repository.SettingRepo
	taskRepo       repository.TaskRepo

	clusterService clusterservice.Service
	traefikService traefikservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}
