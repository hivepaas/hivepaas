package registryuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc"
)

type UC struct {
	*settings.BaseUC

	settingRepo repository.SettingRepo

	registryService registryservice.Service
	registryAuthUC  *registryauthuc.UC
	taskQueue       queue.TaskQueue
}

func New(
	baseUC *settings.BaseUC,

	settingRepo repository.SettingRepo,

	registryService registryservice.Service,
	registryAuthUC *registryauthuc.UC,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		BaseUC: baseUC,

		settingRepo: settingRepo,

		registryService: registryService,
		registryAuthUC:  registryAuthUC,
		taskQueue:       taskQueue,
	}
}
