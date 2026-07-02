package agentserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,
) agentservice.Service {
	return &service{
		dockerManager: dockerManager,
	}
}

type service struct {
	dockerManager docker.Manager
}
