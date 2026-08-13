package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
)

func InitSystemEventBus(
	lc fx.Lifecycle,
	eventBus cacherepository.SystemEventBus,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting system event bus...")
			return eventBus.Start(context.Background()) //nolint:contextcheck
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping system event bus...")
			return eventBus.Stop()
		},
	})
}
