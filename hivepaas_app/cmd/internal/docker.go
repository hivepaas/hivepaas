package internal

import (
	"context"
	"time"

	"github.com/moby/moby/client"
	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	dockerEventsRetryInterval    = 3 * time.Second
	dockerEventsRetryIncr        = 3 * time.Second
	dockerEventsRetryIntervalMax = 5 * 60 * time.Second
)

func InitDockerManager(
	lc fx.Lifecycle,
	cfg *config.Config,
	manager docker.Manager,
	clusterService clusterservice.Service,
	logger logging.Logger,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			logger.Info("initializing docker manager ...")
			if cfg.RunMode == config.RunModeApp || cfg.RunMode == config.RunModeAppAndWorker {
				go registerSwarmNodeEvents(ctx, manager, clusterService, logger)
			}
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			logger.Info("closing docker manager ...")
			cancel()
			return manager.Close()
		},
	})
	return nil
}

//nolint:gocognit
func registerSwarmNodeEvents(
	ctx context.Context,
	manager docker.Manager,
	clusterService clusterservice.Service,
	logger logging.Logger,
) {
	retryInterval := dockerEventsRetryInterval
	for {
		if ctx.Err() != nil {
			return
		}

		res := manager.Events(ctx, func(opts *client.EventsListOptions) {
			docker.FilterAdd(&opts.Filters, "type", "node")
			docker.FilterAdd(&opts.Filters, "event", "create")
		})
		logger.Info("successfully registered swarm node events listener")

		errStreamClosed := false
		for !errStreamClosed {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-res.Messages:
				if !ok {
					logger.Warn("swarm node events channel closed, reconnecting...")
					errStreamClosed = true
					break
				}
				if err := clusterService.OnNodeEvent(ctx, &msg); err != nil {
					logger.Error("failed to process node event", "error", err)
				}
			case err, ok := <-res.Err:
				if !ok {
					logger.Warn("swarm node events error channel closed, reconnecting...")
					errStreamClosed = true
					break
				}
				if err != nil {
					logger.Error("docker event stream error", "error", err)
					errStreamClosed = true
					break
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}

		retryInterval += dockerEventsRetryIncr
		if retryInterval > dockerEventsRetryIntervalMax {
			retryInterval = dockerEventsRetryIntervalMax
		}
	}
}
