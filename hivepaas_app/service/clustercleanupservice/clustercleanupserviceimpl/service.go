package clustercleanupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dockerManager docker.Manager
}

func New(
	dockerManager docker.Manager,
) clustercleanupservice.Service {
	return &service{
		dockerManager: dockerManager,
	}
}
