package loggingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
	logsvc "github.com/hivepaas/hivepaas/services/logging"
)

type service struct {
	dockerManager    docker.Manager
	settingRepo      repository.SettingRepo
	systemAppService systemappservice.Service
	logger           logging.Logger

	// newBackend builds a backend client; a field so tests can substitute one.
	newBackend func(logsvc.BackendType, *logsvc.BackendConfig) (logsvc.Backend, error)
}

// New builds the logging service. fx wires the arguments from the provider list
// in registry/provides.go; adding a parameter here needs no other change.
func New(
	dockerManager docker.Manager,
	settingRepo repository.SettingRepo,
	systemAppService systemappservice.Service,
	logger logging.Logger,
) loggingservice.Service {
	return &service{
		dockerManager:    dockerManager,
		settingRepo:      settingRepo,
		systemAppService: systemAppService,
		logger:           logger,
		newBackend:       logsvc.NewBackend,
	}
}
