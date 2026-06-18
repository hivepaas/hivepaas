package repocheckoutserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/repocheckoutservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	logger         logging.Logger
	db             *database.DB
	redisClient    rediscache.Client
	redisLock      rediscache.Lock
	settingRepo    repository.SettingRepo
	fileRepo       repository.FileRepo
	dockerManager  docker.Manager
	settingService settingservice.Service
}

func New(
	logger logging.Logger,
	db *database.DB,
	redisClient rediscache.Client,
	redisLock rediscache.Lock,
	settingRepo repository.SettingRepo,
	fileRepo repository.FileRepo,
	dockerManager docker.Manager,
	settingService settingservice.Service,
) repocheckoutservice.Service {
	return &service{
		logger:         logger,
		db:             db,
		redisClient:    redisClient,
		redisLock:      redisLock,
		settingRepo:    settingRepo,
		fileRepo:       fileRepo,
		dockerManager:  dockerManager,
		settingService: settingService,
	}
}
