package appplacementsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType = base.SettingTypeAppPlacement
)

type UC struct {
	dockerManager docker.Manager

	*settings.BaseUC
}

func New(
	dockerManager docker.Manager,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		dockerManager: dockerManager,

		BaseUC: baseUC,
	}
}
