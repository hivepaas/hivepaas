package internal

import (
	"context"
	"log"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
)

func InitLogger(lc fx.Lifecycle, cfg *config.Config, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("initializing logger...")
			logging.InitGlobalLogger(logger)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if zapLogger, ok := logger.(*logging.ZapLogger); ok {
				// Ensure that the logger is properly synced before shutdown.
				// Error from sync method is safe to be ignored, but we logged it anyway.
				if err := zapLogger.Sync(); err != nil {
					log.Printf("zap sync warning: %+v\n", err)
				}
			}
			return nil
		},
	})
}
