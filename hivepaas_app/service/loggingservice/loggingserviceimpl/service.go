package loggingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dockerManager  docker.Manager
	clusterService clusterservice.Service
	settingRepo    repository.SettingRepo
	logger         logging.Logger
}

// New builds the logging service. fx wires the arguments from the provider list
// in registry/provides.go; adding a parameter here needs no other change.
func New(
	dockerManager docker.Manager,
	clusterService clusterservice.Service,
	settingRepo repository.SettingRepo,
	logger logging.Logger,
) loggingservice.Service {
	return &service{
		dockerManager:  dockerManager,
		clusterService: clusterService,
		settingRepo:    settingRepo,
		logger:         logger,
	}
}
