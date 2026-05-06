package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tracerr"
	"github.com/localpaas/localpaas/localpaas_app/updater"
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
