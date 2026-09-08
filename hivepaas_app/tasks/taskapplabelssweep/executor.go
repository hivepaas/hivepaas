// Package taskapplabelssweep pushes a confirmed proxy topology onto the apps that
// were left alone while the change was on trial.
//
// It is a task rather than part of the confirm request because by then nothing is
// guarding it: the trial is over, and a fan-out that fails halfway would leave
// some apps on the new depth and some on the old with nothing left to notice.
// Retries and a durable record are what replace the countdown here.
package taskapplabelssweep

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appRoutingService approutingservice.Service
	logger            logging.Logger
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appRoutingService approutingservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{
		appRoutingService: appRoutingService,
		logger:            logger,
	}
	taskQueue.RegisterExecutor(base.TaskTypeAppLabelsSweep, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) error {
	args, err := task.Task.ArgsAsAppLabelsSweep()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("app labels sweep task has no args")
	}

	// Not PrimaryOnly: this is the fan-out the trial deferred. The HivePaaS app is
	// swept again with the rest, which costs one service update with unchanged
	// labels and heals the case where the apply-time sweep did not stick.
	resp, err := e.appRoutingService.ReapplyClientIPStrategy(ctx, db,
		&approutingservice.ReapplyClientIPStrategyReq{PrimaryAppID: args.AppID})
	if err != nil {
		return hperrors.Wrap(err)
	}

	for appID, reason := range resp.Failed {
		e.logger.Errorf("failed to sweep client IP strategy onto app %s: %s", appID, reason)
	}
	task.Task.MustSetOutput(&entity.TaskAppLabelsSweepOutput{
		Applied: resp.Applied,
		Skipped: resp.Skipped,
		Failed:  resp.Failed,
	})

	// Failures are surfaced through the task rather than swallowed: a retry gets
	// another go at them, and if it runs out the task is left failed for somebody
	// to see.
	if len(resp.Failed) > 0 {
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("client IP strategy could not be applied to %d app(s)", len(resp.Failed))
	}
	return nil
}
