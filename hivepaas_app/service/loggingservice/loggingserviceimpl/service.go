package loggingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
	logsvc "github.com/hivepaas/hivepaas/services/logging"
)

type service struct {
	appRepo        repository.AppRepo
	dockerManager  docker.Manager
	clusterService clusterservice.Service
	settingRepo    repository.SettingRepo
	logger         logging.Logger
	hpAppService   hpappservice.Service
	networkService networkservice.Service

	// newBackend builds a backend client; a field so tests can substitute one.
	newBackend func(logsvc.BackendType, *logsvc.BackendConfig) (logsvc.Backend, error)
}

// New builds the logging service. fx wires the arguments from the provider list
// in registry/provides.go; adding a parameter here needs no other change.
func New(
	appRepo repository.AppRepo,
	dockerManager docker.Manager,
	clusterService clusterservice.Service,
	settingRepo repository.SettingRepo,
	logger logging.Logger,
	hpAppService hpappservice.Service,
	networkService networkservice.Service,
) loggingservice.Service {
	return &service{
		appRepo:        appRepo,
		dockerManager:  dockerManager,
		clusterService: clusterService,
		settingRepo:    settingRepo,
		logger:         logger,
		hpAppService:   hpAppService,
		networkService: networkService,
		newBackend:     logsvc.NewBackend,
	}
}
