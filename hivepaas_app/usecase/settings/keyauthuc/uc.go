package keyauthuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeKeyAuth
	currentSettingVersion = entity.CurrentKeyAuthVersion
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
