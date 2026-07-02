package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/services/docker"
)

func InitDockerManager(lc fx.Lifecycle, manager docker.Manager, logger logging.Logger) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("initializing docker manager ...")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing docker manager ...")
			return manager.Close()
		},
	})
	return nil
}
