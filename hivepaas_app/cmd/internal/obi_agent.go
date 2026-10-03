package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/obiagentuc"
	"github.com/hivepaas/hivepaas/services/docker"
)

// OBIOnAgent runs OBI on the agent's node for the apps that ask for their
// routes and calls, while the logging settings list the node, and writes what
// it measures to the agent's stdout. See obiagentuc.
func OBIOnAgent(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	dockerManager docker.Manager,
	logger logging.Logger,
) {
	if cfg.RunMode != config.RunModeAgent {
		return
	}
	uc := obiagentuc.New(logger, db, settingRepo, appRepo, dockerManager, hostPrefix())
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			safego.Go("obi", func() { uc.Run(ctx) })
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}
