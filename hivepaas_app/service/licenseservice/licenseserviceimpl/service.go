package licenseserviceimpl

import (
	"context"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license/licensekeys"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	db               *database.DB
	dockerManager    docker.Manager
	settingRepo      repository.SettingRepo
	systemStatusRepo repository.SystemStatusRepo
	planUsageRepo    repository.PlanUsageRepo

	// timeNow, trust and allowDev are what tests replace: the clock, the keys
	// license keys are verified against, and whether development keys count.
	timeNow  func() time.Time
	trust    func(allowDev bool) (*license.Trust, error)
	allowDev func() bool

	// mu guards the verification of the stored key, kept between calls.
	mu       sync.Mutex
	verified *verification

	// clockMu guards the latest time seen.
	clockMu sync.Mutex
	clock   clockState
}

// New builds the license service. fx wires the arguments from the provider
// list in registry/provides.go.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(
	db *database.DB,
	dockerManager docker.Manager,
	settingRepo repository.SettingRepo,
	systemStatusRepo repository.SystemStatusRepo,
	planUsageRepo repository.PlanUsageRepo,
) licenseservice.Service {
	return &service{
		db:               db,
		dockerManager:    dockerManager,
		settingRepo:      settingRepo,
		systemStatusRepo: systemStatusRepo,
		planUsageRepo:    planUsageRepo,
		timeNow:          time.Now,
		trust:            licensekeys.Trusted,
		// Only the development environment, which also opens the dev-helper
		// endpoints, accepts a key signed by the committed development keys.
		allowDev: func() bool { return config.Current().IsDevEnv() },
	}
}

func (s *service) InstallationID(ctx context.Context) (string, error) {
	id, err := s.dockerManager.SwarmClusterID(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return id, nil
}
