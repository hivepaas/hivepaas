// Package tasksettingsrevert undoes a settings change that nobody confirmed.
//
// The task is the record of the probation: while it sits in the queue unstarted,
// a change is on trial; canceling it is what "confirm" means, and running it is
// what happens when the caller never came back. See the usecase side in
// usecase/system/settingsprobation.
//
// The work itself lives in settingsrevertservice, because the queue is only one
// of three things that can trigger it - the app arms a timer of its own, and the
// startup reconciler picks up deadlines that passed while the process was down.
package tasksettingsrevert

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsrevertservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	settingsRevertService settingsrevertservice.Service
	logger                logging.Logger
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	settingsRevertService settingsrevertservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{
		settingsRevertService: settingsRevertService,
		logger:                logger,
	}
	taskQueue.RegisterExecutor(base.TaskTypeSettingsRevert, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) error {
	args, err := task.Task.ArgsAsSettingsRevert()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("settings revert task has no args")
	}

	resp, err := e.settingsRevertService.Revert(ctx, db, args)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if resp.Reverted {
		e.logger.Warnf("reverted unconfirmed %s change on app %s, applied at %v",
			args.SettingType, args.AppID, args.AppliedAt)
	} else {
		e.logger.Infof("skipped reverting %s change on app %s: %s",
			args.SettingType, args.AppID, resp.Reason)
	}
	task.Task.MustSetOutput(&entity.TaskSettingsRevertOutput{
		Reverted: resp.Reverted,
		Reason:   resp.Reason,
	})

	return nil
}
