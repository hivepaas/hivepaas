package appdeploymentuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db                   *database.DB
	redisClient          rediscache.Client
	projectRepo          repository.ProjectRepo
	appRepo              repository.AppRepo
	deploymentRepo       repository.DeploymentRepo
	taskLogRepo          repository.TaskLogRepo
	deploymentInfoRepo   cacherepository.DeploymentInfoRepo
	taskControlRepo      cacherepository.TaskControlRepo
	userService          userservice.Service
	appService           appservice.Service
	appDeploymentService appdeploymentservice.Service
	taskService          taskservice.Service
	taskQueue            queue.TaskQueue
	dockerManager        docker.Manager
}

func New(
	db *database.DB,
	redisClient rediscache.Client,
	projectRepo repository.ProjectRepo,
	appRepo repository.AppRepo,
	deploymentRepo repository.DeploymentRepo,
	taskLogRepo repository.TaskLogRepo,
	deploymentInfoRepo cacherepository.DeploymentInfoRepo,
	taskControlRepo cacherepository.TaskControlRepo,
	userService userservice.Service,
	appService appservice.Service,
	appDeploymentService appdeploymentservice.Service,
	taskService taskservice.Service,
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:                   db,
		redisClient:          redisClient,
		projectRepo:          projectRepo,
		appRepo:              appRepo,
		deploymentRepo:       deploymentRepo,
		taskLogRepo:          taskLogRepo,
		deploymentInfoRepo:   deploymentInfoRepo,
		taskControlRepo:      taskControlRepo,
		userService:          userService,
		appService:           appService,
		appDeploymentService: appDeploymentService,
		taskService:          taskService,
		taskQueue:            taskQueue,
		dockerManager:        dockerManager,
	}
}
