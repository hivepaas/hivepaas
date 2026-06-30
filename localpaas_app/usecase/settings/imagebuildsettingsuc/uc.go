package imagebuildsettingsuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
