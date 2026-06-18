package appdeploymentserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/repocheckoutservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	logger               logging.Logger
	db                   *database.DB
	redisClient          rediscache.Client
	redisLock            rediscache.Lock
	lockRepo             repository.LockRepo
	settingRepo          repository.SettingRepo
	deploymentRepo       repository.DeploymentRepo
	taskLogRepo          repository.TaskLogRepo
	taskRepo             repository.TaskRepo
	fileRepo             repository.FileRepo
	taskInfoRepo         cacherepository.TaskInfoRepo
	deploymentInfoRepo   cacherepository.DeploymentInfoRepo
	dockerManager        docker.Manager
	containerExecService containerexecservice.Service
	repoCheckoutService  repocheckoutservice.Service
	envVarService        envvarservice.Service
	settingService       settingservice.Service
	userService          userservice.Service
	notificationService  notificationservice.Service
}

func New(
	logger logging.Logger,
	db *database.DB,
	redisClient rediscache.Client,
	redisLock rediscache.Lock,
	lockRepo repository.LockRepo,
	settingRepo repository.SettingRepo,
	deploymentRepo repository.DeploymentRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,
	fileRepo repository.FileRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	dockerManager docker.Manager,
	containerExecService containerexecservice.Service,
	repoCheckoutService repocheckoutservice.Service,
	envVarService envvarservice.Service,
	settingService settingservice.Service,
	userService userservice.Service,
	notificationService notificationservice.Service,
) appdeploymentservice.Service {
	return &service{
		logger:               logger,
		db:                   db,
		redisClient:          redisClient,
		redisLock:            redisLock,
		lockRepo:             lockRepo,
		settingRepo:          settingRepo,
		deploymentRepo:       deploymentRepo,
		taskLogRepo:          taskLogRepo,
		taskRepo:             taskRepo,
		fileRepo:             fileRepo,
		taskInfoRepo:         taskInfoRepo,
		deploymentInfoRepo:   deploymentInfoRepo,
		dockerManager:        dockerManager,
		containerExecService: containerExecService,
		repoCheckoutService:  repoCheckoutService,
		envVarService:        envVarService,
		settingService:       settingService,
		userService:          userService,
		notificationService:  notificationService,
	}
}
