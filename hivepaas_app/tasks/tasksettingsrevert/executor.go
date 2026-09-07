// Package tasksettingsrevert undoes a settings change that nobody confirmed.
//
// The task is the record of the probation: while it sits in the queue unstarted,
// a change is on trial; canceling it is what "confirm" means, and running it is
// what happens when the caller never came back. See the usecase side in
// usecase/system/hpappsettingsuc/routing_settings_probation.go.
package tasksettingsrevert

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appRoutingService approutingservice.Service
	hpAppService      hpappservice.Service
	logger            logging.Logger
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appRoutingService approutingservice.Service,
	hpAppService hpappservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{
		appRoutingService: appRoutingService,
		hpAppService:      hpAppService,
		logger:            logger,
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

	resp, err := Run(ctx, db, e.appRoutingService, e.hpAppService, args)
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

// Run performs the revert itself.
//
// It is exported because the queue is not the only thing that triggers a revert:
// the main app arms a timer of its own and the startup reconciler picks up
// deadlines that passed while the process was down. All of them come through
// here, and all of them are safe to overlap - the caller holds the task row and
// RevertSettings refuses to act on a setting that has moved on.
func Run(
	ctx context.Context,
	db database.Tx,
	appRoutingService approutingservice.Service,
	hpAppService hpappservice.Service,
	args *entity.TaskSettingsRevertArgs,
) (*approutingservice.RevertSettingsResp, error) {
	app, err := hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", args.SettingType),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := appRoutingService.RevertSettings(ctx, db, &approutingservice.RevertSettingsReq{
		App:          app,
		SettingID:    args.SettingID,
		ProbationVer: args.ProbationVer,
		Snapshot:     args.Snapshot,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}
