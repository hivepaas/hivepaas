package domainsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType = base.SettingTypeDomainSettings
)

type UC struct {
	*settings.BaseUC
}

func New(
	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		BaseUC: baseUC,
	}
}
