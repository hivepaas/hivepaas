package imagebuildsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType = base.SettingTypeImageBuildSettings
)

type UC struct {
	sysCleanupService syscleanupservice.Service

	*settings.BaseUC
}

func New(
	sysCleanupService syscleanupservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		sysCleanupService: sysCleanupService,

		BaseUC: baseUC,
	}
}
