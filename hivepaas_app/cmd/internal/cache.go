package internal

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
)

func InitCache(lc fx.Lifecycle, client rediscache.Client, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("initializing cache...")
			if err := client.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("failed to ping redis: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping cache...")
			return client.Close() //nolint:wrapcheck
		},
	})
}
