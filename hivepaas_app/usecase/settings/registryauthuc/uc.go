package registryauthuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryauthrenewaluc"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType    = base.SettingTypeRegistryAuth
	currentSettingVersion = entity.CurrentRegistryAuthVersion
)

type UC struct {
	*settings.BaseUC
	dockerManager       docker.Manager
	registryAuthService registryauthservice.Service
	// registryAuthRenewalUC renews at once the services of a credential whose
	// AWS keys were saved.
	registryAuthRenewalUC *registryauthrenewaluc.UC
}

func New(
	baseUC *settings.BaseUC,
	dockerManager docker.Manager,
	registryAuthService registryauthservice.Service,
	registryAuthRenewalUC *registryauthrenewaluc.UC,
) *UC {
	return &UC{
		BaseUC:                baseUC,
		dockerManager:         dockerManager,
		registryAuthService:   registryAuthService,
		registryAuthRenewalUC: registryAuthRenewalUC,
	}
}
