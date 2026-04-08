package apikeyuc

import (
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UC struct {
	*settings.BaseUC
}

func New(
	baseSettingUC *settings.BaseUC,
) *UC {
	return &UC{
		BaseUC: baseSettingUC,
	}
}
