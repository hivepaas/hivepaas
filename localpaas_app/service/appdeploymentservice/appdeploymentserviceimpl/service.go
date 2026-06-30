package appdeploymentserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/imagebuildservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/repocheckoutservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/services/docker"
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
