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
	*settings.BaseUC
	sysCleanupService syscleanupservice.Service
}

func New(
	baseUC *settings.BaseUC,
	sysCleanupService syscleanupservice.Service,
) *UC {
	return &UC{
		BaseUC:            baseUC,
		sysCleanupService: sysCleanupService,
	}
}
