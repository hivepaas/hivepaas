package traefikserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
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
