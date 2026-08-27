package nodeexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
)

type service struct {
	agentService agentservice.Service
}

func New(
	agentService agentservice.Service,
) nodeexecservice.Service {
	return &service{
		agentService: agentService,
	}
}
