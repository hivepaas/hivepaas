package appdeploymentserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	db          *database.DB
	redisClient rediscache.Client

	deploymentInfoRepo  cacherepository.DeploymentInfoRepo
	deploymentRepo      repository.DeploymentRepo
	lockRepo            repository.LockRepo
	repoCheckoutService repocheckoutservice.Service
	settingRepo         repository.SettingRepo
	taskLogRepo         repository.TaskLogRepo

	containerExecService containerexecservice.Service
	imageBuildService    imagebuildservice.Service
	notificationService  notificationservice.Service
	settingService       settingservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,
	redisClient rediscache.Client,

	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	deploymentRepo repository.DeploymentRepo,
	lockRepo repository.LockRepo,
	repoCheckoutService repocheckoutservice.Service,
	settingRepo repository.SettingRepo,
	taskLogRepo repository.TaskLogRepo,

	containerExecService containerexecservice.Service,
	imageBuildService imagebuildservice.Service,
	notificationService notificationservice.Service,
	settingService settingservice.Service,

	dockerManager docker.Manager,
) appdeploymentservice.Service {
	return &service{
		db:          db,
		redisClient: redisClient,

		deploymentInfoRepo:  deploymentInfoRepo,
		deploymentRepo:      deploymentRepo,
		lockRepo:            lockRepo,
		repoCheckoutService: repoCheckoutService,
		settingRepo:         settingRepo,
		taskLogRepo:         taskLogRepo,

		containerExecService: containerExecService,
		imageBuildService:    imageBuildService,
		notificationService:  notificationService,
		settingService:       settingService,

		dockerManager: dockerManager,
	}
}
