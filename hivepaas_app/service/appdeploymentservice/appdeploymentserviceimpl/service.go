package appdeploymentserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	db            *database.DB
	redisClient   rediscache.Client
	dockerManager docker.Manager

	deploymentInfoRepo cacherepository.DeploymentInfoRepo
	deploymentRepo     repository.DeploymentRepo
	lockRepo           repository.LockRepo
	settingRepo        repository.SettingRepo
	taskLogRepo        repository.TaskLogRepo

	agentService         agentservice.Service
	appService           appservice.Service
	clusterService       clusterservice.Service
	containerExecService containerexecservice.Service
	imageBuildService    imagebuildservice.Service
	notificationService  notificationservice.Service
	placementService     placementservice.Service
	repoCheckoutService  repocheckoutservice.Service
	settingService       settingservice.Service
}

func New(
	db *database.DB,
	redisClient rediscache.Client,
	dockerManager docker.Manager,

	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	deploymentRepo repository.DeploymentRepo,
	lockRepo repository.LockRepo,
	settingRepo repository.SettingRepo,
	taskLogRepo repository.TaskLogRepo,

	agentService agentservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	imageBuildService imagebuildservice.Service,
	notificationService notificationservice.Service,
	placementService placementservice.Service,
	repoCheckoutService repocheckoutservice.Service,
	settingService settingservice.Service,
) appdeploymentservice.Service {
	return &service{
		db:            db,
		redisClient:   redisClient,
		dockerManager: dockerManager,

		deploymentInfoRepo: deploymentInfoRepo,
		deploymentRepo:     deploymentRepo,
		lockRepo:           lockRepo,
		settingRepo:        settingRepo,
		taskLogRepo:        taskLogRepo,

		agentService:         agentService,
		appService:           appService,
		clusterService:       clusterService,
		containerExecService: containerExecService,
		imageBuildService:    imageBuildService,
		notificationService:  notificationService,
		placementService:     placementService,
		repoCheckoutService:  repoCheckoutService,
		settingService:       settingService,
	}
}
