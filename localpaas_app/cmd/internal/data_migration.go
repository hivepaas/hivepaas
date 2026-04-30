package internal

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice"
)

func MigrateData(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	dbService dbservice.Service,
	logger logging.Logger,
) {
	stepEnabled := cfg.RunMode != config.RunModeUpdater
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !stepEnabled {
				return nil
			}
			logger.Info("migrating data structure...")
			if err := dbService.MigrateData(ctx, db); err != nil {
				return fmt.Errorf("failed to migrate data structure: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if !stepEnabled {
				return nil
			}
			return nil
		},
	})
}
