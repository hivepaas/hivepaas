package gitcredentialuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
