package containerexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
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
