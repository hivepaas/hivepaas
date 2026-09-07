package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// ConfirmRoutingSettings vouches for a routing change, which is what stops it
// from being undone.
//
// The proof is the request itself: HivePaaS is reachable only through Traefik, so
// a call that arrives here at all went through the configuration under trial. That
// is the whole of the check, and it is why the settle delay matters - see
// entity.SettingsProbationSettleDelay.
func (uc *UC) ConfirmRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.ConfirmRoutingSettingsReq,
) (*hpappsettingsdto.ConfirmRoutingSettingsResp, error) {
	var confirmed *entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, setting, err := uc.loadRoutingSettingForProbation(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}

		task, args, err := uc.pendingProbationFor(ctx, db, app.ID, setting, req.ChangeID)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if timeutil.NowUTC().Before(args.ConfirmableFrom()) {
			return hperrors.Wrap(hperrors.ErrSettingsConfirmTooEarly).
				WithParam("ConfirmableFrom", args.ConfirmableFrom())
		}

		task.Status = base.TaskStatusCanceled
		task.UpdatedAt = timeutil.NowUTC()
		err = uc.taskRepo.Update(ctx, db, task, bunex.UpdateColumns("status", "updated_at"))
		if err != nil {
			return hperrors.Wrap(err)
		}
		confirmed = task

		return uc.recordProbationOutcome(ctx, db, auth, base.AuditLogTypeRoutingChangeConfirm, setting)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Best effort: the row already says canceled, and the executor refuses to run
	// a task that is not in the not-started state, so a missed unschedule costs a
	// wasted wake-up rather than an unwanted revert.
	if err := uc.taskQueue.UnscheduleTask(ctx, confirmed); err != nil {
		uc.logger.Warnf("failed to unschedule confirmed routing probation %s: %v", confirmed.ID, err)
	}

	return &hpappsettingsdto.ConfirmRoutingSettingsResp{}, nil
}

// RevertRoutingSettings undoes the change now instead of waiting for its deadline.
//
// This is for the caller who can still get in and can see the change was wrong.
// The caller who cannot get in is served by the deadline, which needs nobody.
func (uc *UC) RevertRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.RevertRoutingSettingsReq,
) (*hpappsettingsdto.RevertRoutingSettingsResp, error) {
	var (
		reverted *entity.Task
		output   *entity.TaskSettingsRevertOutput
	)
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, setting, err := uc.loadRoutingSettingForProbation(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}

		task, _, err := uc.pendingProbationFor(ctx, db, app.ID, setting, req.ChangeID)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if err = uc.executeProbationTask(ctx, db, task); err != nil {
			return hperrors.Wrap(err)
		}
		reverted = task
		if output, err = task.OutputAsSettingsRevert(); err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordProbationOutcome(ctx, db, auth, base.AuditLogTypeRoutingChangeRevert, setting)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := uc.taskQueue.UnscheduleTask(ctx, reverted); err != nil {
		uc.logger.Warnf("failed to unschedule reverted routing probation %s: %v", reverted.ID, err)
	}

	resp := &hpappsettingsdto.RevertRoutingSettingsResp{Data: &hpappsettingsdto.RevertRoutingDataResp{}}
	if output != nil {
		resp.Data.Reverted = output.Reverted
		resp.Data.Reason = output.Reason
	}
	return resp, nil
}

// loadRoutingSettingForProbation takes the same lock the update path takes.
//
// Locking the setting rather than the task is what serializes confirm, revert and
// a concurrent update against each other - see findPendingProbation.
func (uc *UC) loadRoutingSettingForProbation(
	ctx context.Context,
	db database.Tx,
) (*entity.App, *entity.Setting, error) {
	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	setting := app.GetSettingByType(base.SettingTypeAppRouting)
	if setting == nil {
		return nil, nil, hperrors.Wrap(hperrors.ErrSettingsNoPendingChange)
	}
	return app, setting, nil
}

// pendingProbationFor returns the trial the caller is talking about, refusing
// anything that has already been overtaken.
func (uc *UC) pendingProbationFor(
	ctx context.Context,
	db database.Tx,
	appID string,
	setting *entity.Setting,
	changeID string,
) (*entity.Task, *entity.TaskSettingsRevertArgs, error) {
	task, err := uc.findPendingProbation(ctx, db, appID)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if task == nil {
		return nil, nil, hperrors.Wrap(hperrors.ErrSettingsNoPendingChange)
	}
	// A confirmation that names a different change was written for a state of the
	// world that has moved on. Accepting it would vouch for settings its sender
	// never saw.
	if changeID != "" && changeID != task.ID {
		return nil, nil, hperrors.Wrap(hperrors.ErrSettingsChangeSuperseded)
	}

	args, err := task.ArgsAsSettingsRevert()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if args == nil {
		return nil, nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("routing probation task %s has no args", task.ID)
	}
	if args.ProbationVer != setting.UpdateVer {
		return nil, nil, hperrors.Wrap(hperrors.ErrSettingsChangeSuperseded)
	}
	return task, args, nil
}
