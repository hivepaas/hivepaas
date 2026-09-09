package settingsprobationserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
)

// Arm implements settingsprobationservice.Service
func (s *service) Arm(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	in *settingsprobationservice.ArmReq,
	out *settingsprobationservice.ArmResult,
	upsertTask func(task *entity.Task),
) error {
	timeNow := timeutil.NowUTC()
	setting := in.Setting

	snapshot := in.Snapshot
	pending, err := s.FindPending(ctx, db, in.AppID, setting.Type)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if pending != nil {
		// Keep the snapshot the pending trial was going to restore, not the state
		// as it stands now. What stands now is the previous unconfirmed change,
		// which may well be the one that locked the caller out - reverting to it
		// would undo nothing. Repeated blind attempts all fall back to the last
		// configuration somebody actually vouched for.
		if prevArgs, e := pending.ArgsAsSettingsRevert(); e == nil && prevArgs != nil &&
			prevArgs.Snapshot.Data != "" {
			snapshot = prevArgs.Snapshot
		}
		pending.Status = base.TaskStatusCanceled
		pending.UpdatedAt = timeNow
		upsertTask(pending)
		out.SupersededProbation = pending
	}

	// Floored, not taken as given: the settle delay is a correctness bound before
	// it is a caller's preference, and a caller that passes nothing gets it.
	confirmableAt := timeNow.Add(max(in.SettleDelay, entity.SettingsProbationSettleDelay))
	deadlineAt := timeNow.Add(in.Window)
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: in.AppID,
		Type:     base.TaskTypeSettingsRevert,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityCritical,
			MaxRetry:   maxRetry,
			RetryDelay: timeutil.Duration(retryDelay),
			Timeout:    timeutil.Duration(taskTimeout),
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     deadlineAt,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	err = task.SetArgs(&entity.TaskSettingsRevertArgs{
		AppID:         in.AppID,
		SettingID:     setting.ID,
		SettingType:   setting.Type,
		ProbationVer:  setting.UpdateVer,
		Snapshot:      snapshot,
		AppliedBy:     auth.UserID(),
		AppliedAt:     timeNow,
		ConfirmableAt: confirmableAt,
		DeadlineAt:    deadlineAt,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	upsertTask(task)
	out.Probation = task
	return nil
}

// Schedule implements settingsprobationservice.Service
//
// Three triggers, one execution. The queue is the normal one; the in-process
// timer covers a worker that is not running at all, which is the default shape
// once RunWorkerInMainApp is turned off; the startup reconciler covers a deadline
// that passed while the process was down. They are safe to overlap because the
// task row is claimed with FOR UPDATE SKIP LOCKED and only in the not-started
// state, so whoever gets there first is the only one that acts.
func (s *service) Schedule(ctx context.Context, out *settingsprobationservice.ArmResult) {
	if out == nil {
		return
	}
	if out.SupersededProbation != nil {
		if err := s.taskQueue.UnscheduleTask(ctx, out.SupersededProbation); err != nil {
			s.logger.Warnf("failed to unschedule superseded settings probation %s: %v",
				out.SupersededProbation.ID, err)
		}
	}
	if out.Probation == nil {
		return
	}
	if err := s.taskQueue.ScheduleTask(ctx, out.Probation); err != nil {
		s.logger.Warnf("failed to schedule settings probation %s, falling back to the local timer "+
			"and the startup scan: %v", out.Probation.ID, err)
	}
	s.armFallback(out.Probation) //nolint:contextcheck // the timer outlives this request
}

// armFallback runs the revert from this process if nothing else did.
//
// The main app is the one place the revert is certain to be able to run: none of
// the changes put on trial recreate its task, so the process that published the
// change is still alive to undo it, and it reaches docker over the socket rather
// than through traefik. The worker may be a separate service with zero replicas.
func (s *service) armFallback(task *entity.Task) {
	delay := max(time.Until(task.RunAt)+fallbackLag, 0)
	taskID := task.ID
	time.AfterFunc(delay, func() {
		defer safego.RecoverWithLogger(s.logger, "settingsprobation.fallback")
		if err := s.runFallback(taskID); err != nil {
			s.logger.Errorf("settings probation fallback failed for task %s: %v", taskID, err)
		}
	})
}

// runFallback claims the task and reverts, if nothing else got there.
//
// The context is a fresh background one on purpose: the request that armed this
// returned minutes ago, and its context is long canceled. Inheriting it would
// cancel the revert exactly when it is most needed.
//
//nolint:contextcheck // deliberately detached from the request that armed it
func (s *service) runFallback(taskID string) error {
	ctx := context.Background()
	return hperrors.Wrap(transaction.Execute(ctx, s.db, func(db database.Tx) error {
		task, err := s.taskRepo.GetByID(ctx, db, nil, base.TaskTypeSettingsRevert, taskID,
			bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
			bunex.SelectFor("UPDATE OF task SKIP LOCKED"),
		)
		if err != nil {
			// Confirmed, already reverted, or held by the worker right now. All
			// three mean somebody else owns the outcome.
			if errors.Is(err, hperrors.ErrNotFound) {
				return nil
			}
			return hperrors.Wrap(err)
		}
		return s.execute(ctx, db, task)
	}))
}

// execute runs the revert and closes the task out.
//
// The caller must already hold the task row. Used by the local fallback, by
// RevertNow and by the reconciler; the queue reaches the same work through the
// task executor, which does its own claiming.
func (s *service) execute(ctx context.Context, db database.Tx, task *entity.Task) error {
	args, err := task.ArgsAsSettingsRevert()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("settings probation task %s has no args", task.ID)
	}

	resp, err := s.settingsRevertService.Revert(ctx, db, args)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if resp.Reverted {
		s.logger.Warnf("reverted unconfirmed %s change on app %s, applied at %v",
			args.SettingType, args.AppID, args.AppliedAt)
	}

	timeNow := timeutil.NowUTC()
	task.Status = base.TaskStatusDone
	task.StartedAt = timeNow
	task.EndedAt = timeNow
	task.UpdatedAt = timeNow
	task.MustSetOutput(&entity.TaskSettingsRevertOutput{Reverted: resp.Reverted, Reason: resp.Reason})
	return hperrors.Wrap(s.taskRepo.Update(ctx, db, task))
}

// FindPending implements settingsprobationservice.Service
//
// The setting type is not a column - it lives in the task args - so it is matched
// here rather than in the query. Matching it at all is the point: two kinds of
// settings can be on trial at the same time, and they are not interchangeable.
// Treating them as one would let a confirmation for one vouch for the other, and
// would let Arm carry a routing snapshot into a service settings task, whose
// revert would then write routing JSON over the service settings row.
//
// No row lock is taken here. Every writer reaches this after locking the setting
// it is about, so that row is what serializes them; locking the task as well
// would add a second order to acquire and nothing else.
func (s *service) FindPending(
	ctx context.Context,
	db database.IDB,
	appID string,
	settingType base.SettingType,
) (*entity.Task, error) {
	tasks, _, err := s.taskRepo.ListByTarget(ctx, db, "", nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSettingsRevert),
		bunex.SelectWhere("task.object_id = ?", appID),
		bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var latest *entity.Task
	for _, task := range tasks {
		args, e := task.ArgsAsSettingsRevert()
		if e != nil || args == nil || args.SettingType != settingType {
			continue
		}
		if latest == nil || task.CreatedAt.After(latest.CreatedAt) {
			latest = task
		}
	}
	return latest, nil
}

// Reconcile implements settingsprobationservice.Service
//
// Called at startup, by the process that serves the API - the one none of these
// changes can take down.
func (s *service) Reconcile(ctx context.Context) error {
	tasks, _, err := s.taskRepo.ListByTarget(ctx, s.db, "", nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSettingsRevert),
		bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	for _, task := range tasks {
		if task.RunAt.After(timeNow) {
			s.armFallback(task) //nolint:contextcheck // the timer outlives this call
			continue
		}
		s.logger.Warnf("settings probation %s ran out at %v with nothing to enforce it, reverting now",
			task.ID, task.RunAt)
		if err := s.runFallback(task.ID); err != nil { //nolint:contextcheck // detached on purpose
			// One overdue trial failing is not a reason to abandon the rest, and
			// definitely not a reason to fail startup.
			s.logger.Errorf("failed to revert overdue settings probation %s: %v", task.ID, err)
		}
	}
	return nil
}
