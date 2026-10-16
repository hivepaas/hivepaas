package oauthuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeOAuth
	currentSettingVersion = entity.CurrentOAuthVersion
)

type UC struct {
	licenseService licenseservice.Service

	*settings.BaseUC
}

func New(
	licenseService licenseservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		licenseService: licenseService,

		BaseUC: baseUC,
	}
}
