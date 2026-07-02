package networkserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
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
