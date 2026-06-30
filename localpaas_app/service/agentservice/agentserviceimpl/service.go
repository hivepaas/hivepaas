package agentserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/agentservice"
	"github.com/localpaas/localpaas/services/docker"
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
