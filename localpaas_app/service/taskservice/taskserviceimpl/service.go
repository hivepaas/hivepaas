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
	taskRepo repository.TaskRepo,
	taskLogRepo repository.TaskLogRepo,
	settingRepo repository.SettingRepo,
	lockRepo repository.LockRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	taskControlRepo cacherepository.TaskControlRepo,
) taskservice.Service {
	return &service{
		db:              db,
		redisClient:     redisClient,
		taskRepo:        taskRepo,
		taskLogRepo:     taskLogRepo,
		settingRepo:     settingRepo,
		lockRepo:        lockRepo,
		taskInfoRepo:    taskInfoRepo,
		taskControlRepo: taskControlRepo,
	}
}

type service struct {
	db              *database.DB
	redisClient     rediscache.Client
	taskRepo        repository.TaskRepo
	taskLogRepo     repository.TaskLogRepo
	settingRepo     repository.SettingRepo
	lockRepo        repository.LockRepo
	taskInfoRepo    cacherepository.TaskInfoRepo
	taskControlRepo cacherepository.TaskControlRepo
}
