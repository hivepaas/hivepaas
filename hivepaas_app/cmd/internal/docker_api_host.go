package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/dockerapiagentuc"
)

// InitDockerAPIHost serves, on this node, the Docker API of every app given it,
// and removes what their children leave behind. It runs in the agent, the one
// process on every node that holds the node's Docker socket.
func InitDockerAPIHost(lc fx.Lifecycle, uc *dockerapiagentuc.UC, logger logging.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			safego.GoWithLogger(logger, "dockerAPIHost", func() {
				uc.Run(ctx)
			})
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			uc.Close()
			return nil
		},
	})
}
