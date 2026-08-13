package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
)

func InitSystemEventBus(
	lc fx.Lifecycle,
	eventBusService systemeventbusservice.Service,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting system event bus...")
			return eventBusService.Start(context.Background()) //nolint:contextcheck
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping system event bus...")
			return eventBusService.Stop()
		},
	})
}
