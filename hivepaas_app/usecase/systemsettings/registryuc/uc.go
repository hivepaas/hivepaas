package registryuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UC struct {
	*settings.BaseUC

	settingRepo repository.SettingRepo

	registryService registryservice.Service
}

func New(
	baseUC *settings.BaseUC,

	settingRepo repository.SettingRepo,

	registryService registryservice.Service,
) *UC {
	return &UC{
		BaseUC: baseUC,

		settingRepo: settingRepo,

		registryService: registryService,
	}
}
