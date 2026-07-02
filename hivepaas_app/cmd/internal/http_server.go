package internal

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/server"
)

func InitHTTPServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	srv server.Server,
	logger logging.Logger,
) {
	stepEnabled := cfg.RunMode == config.RunModeApp || cfg.RunMode == config.RunModeAppAndWorker
	if !stepEnabled {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("starting HTTP server ...")
			go func() {
				if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Fatalf("start server error: %v", err.Error())
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping HTTP server ...")
			return srv.Stop(ctx)
		},
	})
}
