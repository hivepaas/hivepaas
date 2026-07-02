package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

func InitJWTSession(lc fx.Lifecycle, cfg *config.Config, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("initializing JWT session...")
			return jwtsession.InitJWTSession(&jwtsession.Config{
				Secret:          cfg.Session.JWTSecret,
				AccessTokenExp:  cfg.Session.AccessTokenExp,
				RefreshTokenExp: cfg.Session.RefreshTokenExp,
			}) //nolint:wrapcheck
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping JWT session...")
			return nil
		},
	})
}
