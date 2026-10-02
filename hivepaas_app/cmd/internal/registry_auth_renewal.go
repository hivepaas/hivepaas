package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryauthrenewaluc"
)

// RegistryAuthRenewalOnStart runs the registry auth renewal at start when its
// last run is older than its interval: while HivePaaS was down, the Amazon ECR
// tokens Swarm keeps went on aging. It runs once the task queue has started,
// so the task is scheduled at once. A failure is only logged.
func RegistryAuthRenewalOnStart(
	lc fx.Lifecycle,
	cfg *config.Config,
	renewalUC *registryauthrenewaluc.UC,
	logger logging.Logger,
) {
	if cfg.RunMode == config.RunModeUpdater {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			scheduled, err := renewalUC.RenewIfStale(ctx)
			if err != nil {
				logger.Errorf("failed to check the registry auth renewal: %v", err)
				return nil
			}
			if scheduled {
				logger.Info("the registry auth renewal is late: it runs now")
			}
			return nil
		},
	})
}
