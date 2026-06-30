package containerexecserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/agentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	agentService agentservice.Service

	dockerManager docker.Manager
}

func New(
	agentService agentservice.Service,

	dockerManager docker.Manager,
) containerexecservice.Service {
	return &service{
		agentService: agentService,

		dockerManager: dockerManager,
	}
}
