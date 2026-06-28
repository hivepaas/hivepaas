package appserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	db *database.DB,
	appRepo repository.AppRepo,
	appTagRepo repository.AppTagRepo,
	settingRepo repository.SettingRepo,
	resLinkRepo repository.ResLinkRepo,
	deploymentRepo repository.DeploymentRepo,
	taskRepo repository.TaskRepo,
	fileRepo repository.FileRepo,
	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	permissionManager permission.Manager,
	userService userservice.Service,
	traefikService traefikservice.Service,
	clusterService clusterservice.Service,
	dockerManager docker.Manager,
) appservice.Service {
	return &service{
		db:                 db,
		appRepo:            appRepo,
		appTagRepo:         appTagRepo,
		settingRepo:        settingRepo,
		resLinkRepo:        resLinkRepo,
		deploymentRepo:     deploymentRepo,
		taskRepo:           taskRepo,
		fileRepo:           fileRepo,
		deploymentInfoRepo: deploymentInfoRepo,
		permissionManager:  permissionManager,
		userService:        userService,
		traefikService:     traefikService,
		clusterService:     clusterService,
		dockerManager:      dockerManager,
	}
}

type service struct {
	db                 *database.DB
	appRepo            repository.AppRepo
	appTagRepo         repository.AppTagRepo
	settingRepo        repository.SettingRepo
	resLinkRepo        repository.ResLinkRepo
	deploymentRepo     repository.DeploymentRepo
	taskRepo           repository.TaskRepo
	fileRepo           repository.FileRepo
	deploymentInfoRepo cacherepository.DeploymentInfoRepo
	permissionManager  permission.Manager
	userService        userservice.Service
	traefikService     traefikservice.Service
	clusterService     clusterservice.Service
	dockerManager      docker.Manager
}
