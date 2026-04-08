package notificationuc

import (
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/services/docker"
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
