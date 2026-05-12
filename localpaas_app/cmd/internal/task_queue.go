package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/tasks/initializer"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

func InitTaskQueue(
	lc fx.Lifecycle,
	taskQueue queue.TaskQueue,
	_ *initializer.WorkerInitializer,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infof("initializing task queue...")
			if err := taskQueue.Start(); err != nil {
				logger.Errorf("failed to initialize task queue: %v", err)
				return apperrors.Wrap(err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping task queue ...")
			if err := taskQueue.Shutdown(); err != nil {
				logger.Errorf("failed to stop task queue: %v", err)
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
}
