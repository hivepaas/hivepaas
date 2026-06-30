package taskserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
)

func New(
	db *database.DB,
	redisClient rediscache.Client,

	lockRepo repository.LockRepo,
	taskControlRepo cacherepository.TaskControlRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,
) taskservice.Service {
	return &service{
		db:          db,
		redisClient: redisClient,

		lockRepo:        lockRepo,
		taskControlRepo: taskControlRepo,
		taskInfoRepo:    taskInfoRepo,
		taskLogRepo:     taskLogRepo,
		taskRepo:        taskRepo,
	}
}

type service struct {
	db          *database.DB
	redisClient rediscache.Client

	lockRepo        repository.LockRepo
	taskControlRepo cacherepository.TaskControlRepo
	taskInfoRepo    cacherepository.TaskInfoRepo
	taskLogRepo     repository.TaskLogRepo
	taskRepo        repository.TaskRepo
}
