package registryauthuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType    = base.SettingTypeRegistryAuth
	currentSettingVersion = entity.CurrentRegistryAuthVersion
)

type UC struct {
	*settings.BaseUC
	dockerManager docker.Manager
}

func New(
	baseUC *settings.BaseUC,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		BaseUC:        baseUC,
		dockerManager: dockerManager,
	}
}
