package imagebuildsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType = base.SettingTypeImageBuild
)

type UC struct {
	sysCleanupService syscleanupservice.Service
	dockerManager     docker.Manager

	*settings.BaseUC
}

func New(
	sysCleanupService syscleanupservice.Service,
	dockerManager docker.Manager,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		sysCleanupService: sysCleanupService,
		dockerManager:     dockerManager,

		BaseUC: baseUC,
	}
}
