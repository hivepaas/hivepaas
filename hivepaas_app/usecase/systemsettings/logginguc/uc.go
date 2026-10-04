package logginguc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	*settings.BaseUC

	settingRepo   repository.SettingRepo
	obiSettings   cacherepository.OBISettingsRepo
	dockerManager docker.Manager

	loggingService loggingservice.Service
	settingService settingservice.Service
	taskQueue      queue.TaskQueue
}

func New(
	baseUC *settings.BaseUC,

	settingRepo repository.SettingRepo,
	obiSettings cacherepository.OBISettingsRepo,
	dockerManager docker.Manager,

	loggingService loggingservice.Service,
	settingService settingservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		BaseUC: baseUC,

		settingRepo:   settingRepo,
		obiSettings:   obiSettings,
		dockerManager: dockerManager,

		loggingService: loggingService,
		settingService: settingService,
		taskQueue:      taskQueue,
	}
}

// forgetOBISettings drops the settings the agents cache to run OBI, once a
// change to the logging settings is committed: whether the logs are stored,
// and the feature's switch and nodes, are among them. A failure costs at most
// the agents' next read of the database, within 10 minutes.
func (uc *UC) forgetOBISettings(ctx context.Context) {
	_ = uc.obiSettings.Invalidate(context.WithoutCancel(ctx))
}
