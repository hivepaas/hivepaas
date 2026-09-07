package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc"
)

// InitSettingsProbation re-arms confirm-or-revert across a restart.
//
// It runs in the app process rather than the worker on purpose. A routing change
// rewrites swarm service labels, which does not recreate the task, so the app that
// published a change that locked everybody out is still running and is the one
// thing certain to be able to undo it. The worker may be a separate service that
// is scaled to zero.
//
// The sweep is started in the background: nothing else waits on it, and a startup
// that blocks on reaching the database twice is a startup that fails in a new way.
func InitSettingsProbation(
	lc fx.Lifecycle,
	cfg *config.Config,
	hpAppSettingsUC *hpappsettingsuc.UC,
	logger logging.Logger,
) {
	if cfg.RunMode != config.RunModeApp && cfg.RunMode != config.RunModeAppAndWorker {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			//nolint:contextcheck // the sweep outlives the OnStart context, which fx cancels
			safego.Go("reconcileSettingsProbation", func() {
				if err := hpAppSettingsUC.ReconcileProbations(context.Background()); err != nil {
					logger.Errorf("failed to reconcile settings probations: %v", err)
				}
			})
			return nil
		},
	})
}
