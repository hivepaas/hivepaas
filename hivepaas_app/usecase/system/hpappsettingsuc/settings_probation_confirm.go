package hpappsettingsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// ConfirmRoutingSettings vouches for a routing change, which is what stops it
// from being undone.
func (uc *UC) ConfirmRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.ConfirmRoutingSettingsReq,
) (*hpappsettingsdto.ConfirmRoutingSettingsResp, error) {
	if err := uc.confirmSettingsChange(ctx, auth, base.SettingTypeAppRouting, req.ChangeID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.ConfirmRoutingSettingsResp{}, nil
}

// RevertRoutingSettings undoes the change now instead of waiting for its deadline.
func (uc *UC) RevertRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.RevertRoutingSettingsReq,
) (*hpappsettingsdto.RevertRoutingSettingsResp, error) {
	output, err := uc.revertSettingsChange(ctx, auth, base.SettingTypeAppRouting, req.ChangeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.RevertRoutingSettingsResp{Data: transformRevertOutput(output)}, nil
}

// ConfirmServiceSettings vouches for a HivePaaS service settings change.
func (uc *UC) ConfirmServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.ConfirmServiceSettingsReq,
) (*hpappsettingsdto.ConfirmServiceSettingsResp, error) {
	if err := uc.confirmSettingsChange(ctx, auth, base.SettingTypeHivePaaSService, req.ChangeID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.ConfirmServiceSettingsResp{}, nil
}

// RevertServiceSettings undoes an unconfirmed service settings change now.
func (uc *UC) RevertServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.RevertServiceSettingsReq,
) (*hpappsettingsdto.RevertServiceSettingsResp, error) {
	output, err := uc.revertSettingsChange(ctx, auth, base.SettingTypeHivePaaSService, req.ChangeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappsettingsdto.RevertServiceSettingsResp{Data: transformRevertOutput(output)}, nil
}

func transformRevertOutput(output *entity.TaskSettingsRevertOutput) *hpappsettingsdto.RevertSettingsDataResp {
	resp := &hpappsettingsdto.RevertSettingsDataResp{}
	if output != nil {
		resp.Reverted = output.Reverted
		resp.Reason = output.Reason
	}
	return resp
}

// confirmSettingsChange is the whole of the check.
//
// The proof is the request itself: HivePaaS is reachable only through Traefik, so
// a call that arrives here at all traveled through the configuration under trial.
// That is why the settle delay matters - before it, the request would have gone
// through the configuration being replaced. See TaskSettingsRevertArgs.ConfirmableFrom.
func (uc *UC) confirmSettingsChange(
	ctx context.Context,
	auth *basedto.Auth,
	settingType base.SettingType,
	changeID string,
) error {
	var (
		confirmed *entity.Task
		sweep     *entity.Task
		e         error
	)
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		task, args, setting, err := uc.pendingProbationFor(ctx, db, settingType, changeID)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if timeutil.NowUTC().Before(args.ConfirmableFrom()) {
			return hperrors.Wrap(hperrors.ErrSettingsConfirmTooEarly).
				WithParam("ConfirmableFrom", args.ConfirmableFrom())
		}

		task.Status = base.TaskStatusCanceled
		task.UpdatedAt = timeutil.NowUTC()
		if err = uc.taskRepo.Update(ctx, db, task, bunex.UpdateColumns("status", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
		confirmed = task

		// Only the HivePaaS app carried the new proxy topology while the change
		// was on trial. Now that somebody has vouched for it, the rest get it too.
		if sweep, e = uc.scheduleAppLabelsSweep(ctx, db, settingType, args.AppID); e != nil {
			return hperrors.Wrap(e)
		}

		return uc.recordProbationOutcome(ctx, db, auth, base.AuditLogTypeRoutingChangeConfirm, setting)
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Best effort: the row already says canceled, and the executor refuses to run
	// a task that is not in the not-started state, so a missed unschedule costs a
	// wasted wake-up rather than an unwanted revert.
	if err := uc.taskQueue.UnscheduleTask(ctx, confirmed); err != nil {
		uc.logger.Warnf("failed to unschedule confirmed settings probation %s: %v", confirmed.ID, err)
	}
	if sweep != nil {
		// Best effort, as with the probation itself: the row is committed, so the
		// queue's own scan and the startup reconciler will find it even if this
		// call does not land.
		if err := uc.taskQueue.ScheduleTask(ctx, sweep); err != nil {
			uc.logger.Warnf("failed to schedule the app labels sweep %s: %v", sweep.ID, err)
		}
	}

	return nil
}

// scheduleAppLabelsSweep records the fan-out the trial deferred.
//
// Only for the HivePaaS service settings: a routing change is one app's labels
// and has nothing to spread. The task is written in the same transaction as the
// confirmation, so a confirmation that commits always has a sweep to go with it.
func (uc *UC) scheduleAppLabelsSweep(
	ctx context.Context,
	db database.Tx,
	settingType base.SettingType,
	appID string,
) (*entity.Task, error) {
	if settingType != base.SettingTypeHivePaaSService {
		return nil, nil //nolint:nilnil // no sweep is the answer for other setting types
	}

	timeNow := timeutil.NowUTC()
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: appID,
		Type:     base.TaskTypeAppLabelsSweep,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityCritical,
			MaxRetry:   appLabelsSweepMaxRetry,
			RetryDelay: timeutil.Duration(appLabelsSweepRetryDelay),
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err := task.SetArgs(&entity.TaskAppLabelsSweepArgs{AppID: appID}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := uc.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

// revertSettingsChange undoes the change now instead of waiting for its deadline.
//
// This is for the caller who can still get in and can see the change was wrong.
// The caller who cannot get in is served by the deadline, which needs nobody.
func (uc *UC) revertSettingsChange(
	ctx context.Context,
	auth *basedto.Auth,
	settingType base.SettingType,
	changeID string,
) (*entity.TaskSettingsRevertOutput, error) {
	var (
		reverted *entity.Task
		output   *entity.TaskSettingsRevertOutput
	)
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		task, _, setting, err := uc.pendingProbationFor(ctx, db, settingType, changeID)
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
		uc.logger.Warnf("failed to unschedule reverted settings probation %s: %v", reverted.ID, err)
	}

	return output, nil
}

// pendingProbationFor returns the trial the caller is talking about, refusing
// anything that has already been overtaken.
//
// The setting is loaded from the id the task carries rather than from the type
// the endpoint belongs to. That is what lets one piece of code serve both kinds
// of trial, and it is also the safer read: the task is the record of what was put
// on trial, so asking it removes any chance of confirming one setting while
// checking the version of another.
func (uc *UC) pendingProbationFor(
	ctx context.Context,
	db database.Tx,
	settingType base.SettingType,
	changeID string,
) (*entity.Task, *entity.TaskSettingsRevertArgs, *entity.Setting, error) {
	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}

	task, err := uc.findPendingProbation(ctx, db, app.ID, settingType)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	if task == nil {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrSettingsNoPendingChange)
	}
	// A confirmation that names a different change was written for a state of the
	// world that has moved on. Accepting it would vouch for settings its sender
	// never saw.
	if changeID != "" && changeID != task.ID {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrSettingsChangeSuperseded)
	}

	args, err := task.ArgsAsSettingsRevert()
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	if args == nil {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("settings probation task %s has no args", task.ID)
	}

	// Locking the setting, not the task: it is the row every writer of these
	// settings goes through, so it is what serializes confirm, revert and a
	// concurrent update against each other.
	setting, err := uc.settingRepo.GetByID(ctx, db, nil, args.SettingType, args.SettingID, true,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrSettingsNoPendingChange)
	}
	if args.ProbationVer != setting.UpdateVer {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrSettingsChangeSuperseded)
	}

	return task, args, setting, nil
}
