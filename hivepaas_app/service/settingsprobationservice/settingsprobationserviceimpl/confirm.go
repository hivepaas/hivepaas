package settingsprobationserviceimpl

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
)

// Confirm implements settingsprobationservice.Service
//
// The proof is the request itself: HivePaaS is reachable only through traefik, so
// a call that arrives here at all traveled through the configuration under trial.
// That is why the settle delay matters - before it, the request would have gone
// through the configuration being replaced. See TaskSettingsRevertArgs.ConfirmableFrom.
func (s *service) Confirm(ctx context.Context, auth *basedto.Auth, in *settingsprobationservice.AnswerReq) error {
	// Refused rather than skipped. Without the check, a confirmation cancels the
	// trial on the strength of the request having arrived - which after a swarm
	// rollback says nothing about the change it is vouching for. A caller with
	// genuinely nothing to check says so with a function that returns nil, and has
	// to write down why; a caller that simply forgot finds out on the first
	// confirmation rather than on the first rollback.
	if in.EnsureStillLive == nil {
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("confirming a %v change with no liveness check", in.SettingType)
	}

	var (
		confirmed *entity.Task
		followUps []*entity.Task
	)
	err := transaction.Execute(ctx, s.db, func(db database.Tx) error {
		task, args, setting, err := s.pendingFor(ctx, db, in)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if timeutil.NowUTC().Before(args.ConfirmableFrom()) {
			return hperrors.Wrap(hperrors.ErrSettingsConfirmTooEarly).
				WithParam("ConfirmableFrom", args.ConfirmableFrom())
		}

		// Before the task is canceled: canceling is what makes the change
		// permanent, and there is no getting it back afterwards.
		if err = in.EnsureStillLive(ctx, db, args); err != nil {
			return hperrors.Wrap(err)
		}

		task.Status = base.TaskStatusCanceled
		task.UpdatedAt = timeutil.NowUTC()
		if err = s.taskRepo.Update(ctx, db, task, bunex.UpdateColumns("status", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
		confirmed = task

		// Whatever the caller deferred until somebody vouched for the change, in
		// the same transaction as the vouching. See AnswerReq.OnConfirmed.
		if in.OnConfirmed != nil {
			if followUps, err = in.OnConfirmed(ctx, db, args); err != nil {
				return hperrors.Wrap(err)
			}
		}

		return s.recordOutcome(ctx, db, auth, base.AuditLogTypeRoutingChangeConfirm, setting)
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Best effort: the row already says canceled, and the executor refuses to run
	// a task that is not in the not-started state, so a missed unschedule costs a
	// wasted wake-up rather than an unwanted revert.
	if err := s.taskQueue.UnscheduleTask(ctx, confirmed); err != nil {
		s.logger.Warnf("failed to unschedule confirmed settings probation %s: %v", confirmed.ID, err)
	}
	for _, task := range followUps {
		// Best effort, as with the probation itself: the row is committed, so the
		// queue's own scan and the startup reconciler will find it even if this
		// call does not land.
		if err := s.taskQueue.ScheduleTask(ctx, task); err != nil {
			s.logger.Warnf("failed to schedule the follow-up task %s of a confirmed change: %v", task.ID, err)
		}
	}

	return nil
}

// RevertNow implements settingsprobationservice.Service
//
// This is for the caller who can still get in and can see the change was wrong.
// The caller who cannot get in is served by the deadline, which needs nobody.
func (s *service) RevertNow(
	ctx context.Context,
	auth *basedto.Auth,
	in *settingsprobationservice.AnswerReq,
) (*entity.TaskSettingsRevertOutput, error) {
	var (
		reverted *entity.Task
		output   *entity.TaskSettingsRevertOutput
	)
	err := transaction.Execute(ctx, s.db, func(db database.Tx) error {
		task, _, setting, err := s.pendingFor(ctx, db, in)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if err = s.execute(ctx, db, task); err != nil {
			return hperrors.Wrap(err)
		}
		reverted = task
		if output, err = task.OutputAsSettingsRevert(); err != nil {
			return hperrors.Wrap(err)
		}

		return s.recordOutcome(ctx, db, auth, base.AuditLogTypeRoutingChangeRevert, setting)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := s.taskQueue.UnscheduleTask(ctx, reverted); err != nil {
		s.logger.Warnf("failed to unschedule reverted settings probation %s: %v", reverted.ID, err)
	}

	return output, nil
}

// pendingFor returns the trial the caller is talking about, refusing anything
// that has already been overtaken.
//
// The setting is loaded from the id the task carries rather than from the type
// the endpoint belongs to. That is what lets one piece of code serve every kind
// of trial, and it is also the safer read: the task is the record of what was put
// on trial, so asking it removes any chance of confirming one setting while
// checking the version of another.
func (s *service) pendingFor(
	ctx context.Context,
	db database.Tx,
	in *settingsprobationservice.AnswerReq,
) (*entity.Task, *entity.TaskSettingsRevertArgs, *entity.Setting, error) {
	task, err := s.FindPending(ctx, db, in.AppID, in.SettingType)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	if task == nil {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrSettingsNoPendingChange)
	}
	// An answer that names a different change was written for a state of the
	// world that has moved on. Accepting it would vouch for settings its sender
	// never saw.
	if in.ChangeID != "" && in.ChangeID != task.ID {
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
	setting, err := s.settingRepo.GetByID(ctx, db, nil, args.SettingType, args.SettingID, true,
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

func (s *service) recordOutcome(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	typ base.AuditLogType,
	setting *entity.Setting,
) error {
	return hperrors.Wrap(s.auditService.Record(ctx, db, &auditservice.Entry{
		Type:     typ,
		Scope:    setting.Scope,
		ObjectID: setting.ObjectID,
		Source:   base.AuditLogSourceAPIUpdate,
		Result:   base.AuditLogResultAllowed,
		Auth:     auth,
		ResType:  base.ResourceTypeSetting,
		ResID:    setting.ID,
		ResName:  setting.Name,
	}))
}
