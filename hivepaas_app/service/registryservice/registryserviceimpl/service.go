package registryserviceimpl

import (
	"context"
	"net/http"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// httpTimeout is what asking the registry a question is allowed to take. The
// push check is the exception and sets its own.
const httpTimeout = 30 * time.Second

type service struct {
	projectEnvRepo repository.ProjectEnvRepo
	settingRepo    repository.SettingRepo

	systemAppService     systemappservice.Service
	clusterSecretService clustersecretservice.Service

	logger logging.Logger

	// httpClient talks to the registry itself - the status, the domain probe. The
	// push check builds its own, because an upload of that size needs a timeout
	// this one should not have.
	httpClient *http.Client
}

// New builds the registry service. fx wires the arguments from the provider list
// in registry/provides.go; adding a parameter here needs no other change.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(
	projectEnvRepo repository.ProjectEnvRepo,
	settingRepo repository.SettingRepo,

	systemAppService systemappservice.Service,
	clusterSecretService clustersecretservice.Service,

	logger logging.Logger,
) registryservice.Service {
	return &service{
		projectEnvRepo:       projectEnvRepo,
		settingRepo:          settingRepo,
		systemAppService:     systemAppService,
		clusterSecretService: clusterSecretService,
		logger:               logger,
		httpClient:           &http.Client{Timeout: httpTimeout},
	}
}

// Validate refuses a configuration before it is written. It is on the interface so
// that the usecase can refuse in PrepareUpdate, where a refusal is a validation
// error rather than a stored configuration Apply then fails on.
func (s *service) Validate(
	ctx context.Context,
	db database.IDB,
	next, current *entity.RegistrySettings,
) error {
	app, err := s.loadApp(ctx, db)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return validateSettings(next, current, app != nil)
}
