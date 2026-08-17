package containerfileserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	agentService agentservice.Service

	dockerManager docker.Manager
}

func New(
	agentService agentservice.Service,

	dockerManager docker.Manager,
) containerfileservice.Service {
	return &service{
		agentService: agentService,

		dockerManager: dockerManager,
	}
}
