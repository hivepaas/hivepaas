package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
	"github.com/hivepaas/hivepaas/hivepaas_app/updater"
)

func InitUpdater(
	lc fx.Lifecycle,
	cfg *config.Config,
	upd updater.Updater,
	logger logging.Logger,
) {
	if cfg.RunMode != config.RunModeUpdater {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("starting updater ...")
			if err := upd.Start(); err != nil {
				logger.Fatalf("start updater error: %v", err.Error())
				return tracerr.Wrap(err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("stopping updater ...")
			return upd.Shutdown()
		},
	})
}
