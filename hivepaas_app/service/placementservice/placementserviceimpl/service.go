package placementserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,

	settingRepo repository.SettingRepo,

	clusterService clusterservice.Service,
	logger logging.Logger,
) placementservice.Service {
	return &service{
		dockerManager: dockerManager,

		settingRepo: settingRepo,

		clusterService: clusterService,
		logger:         logger,
	}
}

type service struct {
	dockerManager docker.Manager

	settingRepo repository.SettingRepo

	clusterService clusterservice.Service
	logger         logging.Logger
}
