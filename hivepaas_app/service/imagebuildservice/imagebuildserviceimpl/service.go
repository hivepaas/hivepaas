package imagebuildserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	redisClient   rediscache.Client
	redisLock     rediscache.Lock
	dockerManager docker.Manager

	settingRepo repository.SettingRepo

	envVarService  envvarservice.Service
	settingService settingservice.Service
}

func New(
	redisClient rediscache.Client,
	redisLock rediscache.Lock,
	dockerManager docker.Manager,

	settingRepo repository.SettingRepo,

	envVarService envvarservice.Service,
	settingService settingservice.Service,
) imagebuildservice.Service {
	return &service{
		redisClient:   redisClient,
		redisLock:     redisLock,
		dockerManager: dockerManager,

		settingRepo: settingRepo,

		envVarService:  envVarService,
		settingService: settingService,
	}
}
