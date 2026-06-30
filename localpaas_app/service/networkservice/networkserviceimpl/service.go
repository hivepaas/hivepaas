package networkserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	dockerManager docker.Manager,
) networkservice.Service {
	return &service{
		dockerManager: dockerManager,
	}
}

type service struct {
	dockerManager docker.Manager
}
