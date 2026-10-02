package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

// TraefikRouteNamesOnStart renames, in the background, the Traefik routers of
// apps whose services still name them by key: two apps of the same key in two
// projects, or an app named like HivePaaS's own, make Traefik drop the routes
// of both. Labels only, so nothing restarts; an app already renamed is passed
// over, so a start after the first costs a service inspect per app.
func TraefikRouteNamesOnStart(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	appRoutingService approutingservice.Service,
	logger logging.Logger,
) {
	if cfg.RunMode != config.RunModeApp && cfg.RunMode != config.RunModeAppAndWorker {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Past the start's deadline: the sweep outlives it.
			ctx = context.WithoutCancel(ctx)
			safego.Go("traefik route names", func() {
				resp, err := appRoutingService.ReapplyRouteNames(ctx, db)
				if err != nil {
					logger.Errorf("failed to rename the apps' Traefik routers: %v", err)
					return
				}
				if resp.Applied > 0 {
					logger.Infof("renamed the Traefik routers of %d apps by their ids", resp.Applied)
				}
				for appID, reason := range resp.Failed {
					logger.Errorf("failed to rename the Traefik routers of app %s: %s", appID, reason)
				}
			})
			return nil
		},
	})
}
