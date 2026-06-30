package traefikserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/services/docker"
)

func New(
	dockerManager docker.Manager,
) traefikservice.Service {
	return &service{
		dockerManager: dockerManager,
	}
}

type service struct {
	dockerManager docker.Manager
}
