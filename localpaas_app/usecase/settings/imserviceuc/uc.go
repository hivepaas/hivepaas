package imserviceuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeIMService
	currentSettingVersion = entity.CurrentIMServiceVersion
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
