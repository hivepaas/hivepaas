package taskuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
)

type UC struct {
	db              *database.DB
	redisClient     rediscache.Client
	taskRepo        repository.TaskRepo
	taskLogRepo     repository.TaskLogRepo
	taskInfoRepo    cacherepository.TaskInfoRepo
	taskControlRepo cacherepository.TaskControlRepo
	taskService     taskservice.Service
	settingRepo     repository.SettingRepo
}

func New(
	db *database.DB,
	redisClient rediscache.Client,
	taskRepo repository.TaskRepo,
	taskLogRepo repository.TaskLogRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	taskControlRepo cacherepository.TaskControlRepo,
	taskService taskservice.Service,
	settingRepo repository.SettingRepo,
) *UC {
	return &UC{
		db:              db,
		redisClient:     redisClient,
		taskRepo:        taskRepo,
		taskLogRepo:     taskLogRepo,
		taskInfoRepo:    taskInfoRepo,
		taskControlRepo: taskControlRepo,
		taskService:     taskService,
		settingRepo:     settingRepo,
	}
}
