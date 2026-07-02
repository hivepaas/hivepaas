package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
)

func InitDBConnection(lc fx.Lifecycle, db *database.DB, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("pinging db connection...")
			if err := db.Ping(); err != nil {
				logger.Errorf("failed to use connection %v", err.Error())

				return apperrors.New(err)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing db connection...")
			return db.Close()
		},
	})
}
