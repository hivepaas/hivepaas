package appserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	appTagRepo repository.AppTagRepo,
	settingRepo repository.SettingRepo,
	resLinkRepo repository.ResLinkRepo,
	deploymentRepo repository.DeploymentRepo,
	taskRepo repository.TaskRepo,
	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	permissionManager permission.Manager,
	userService userservice.Service,
	traefikService traefikservice.Service,
	dockerManager docker.Manager,
) appservice.Service {
	return &service{
		appRepo:            appRepo,
		appTagRepo:         appTagRepo,
		settingRepo:        settingRepo,
		resLinkRepo:        resLinkRepo,
		deploymentRepo:     deploymentRepo,
		taskRepo:           taskRepo,
		deploymentInfoRepo: deploymentInfoRepo,
		permissionManager:  permissionManager,
		userService:        userService,
		traefikService:     traefikService,
		dockerManager:      dockerManager,
	}
}

type service struct {
	appRepo            repository.AppRepo
	appTagRepo         repository.AppTagRepo
	settingRepo        repository.SettingRepo
	resLinkRepo        repository.ResLinkRepo
	deploymentRepo     repository.DeploymentRepo
	taskRepo           repository.TaskRepo
	deploymentInfoRepo cacherepository.DeploymentInfoRepo
	permissionManager  permission.Manager
	userService        userservice.Service
	traefikService     traefikservice.Service
	dockerManager      docker.Manager
}
