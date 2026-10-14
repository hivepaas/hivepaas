package clustercleanupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dockerManager    docker.Manager
	unusedImagesRepo cacherepository.UnusedImagesRepo
}

func New(
	dockerManager docker.Manager,
	unusedImagesRepo cacherepository.UnusedImagesRepo,
) clustercleanupservice.Service {
	return &service{
		dockerManager:    dockerManager,
		unusedImagesRepo: unusedImagesRepo,
	}
}
