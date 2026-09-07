package hpappsettingsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/tasksettingsrevert"
)

// Confirm-or-revert.
//
// A routing change is applied immediately and then put on trial: unless the
// caller comes back through the new configuration and confirms it, it is undone.
// This is the guard for everything ensureStillReachable cannot see - a broken
// basic-auth reference, a middleware that no longer resolves, a domain whose DNS
// is not ready, a rate limit set too low. None of those can be recognized by
// reading the request; all of them are obvious the moment somebody tries to come
// back in.
//
// It works because a routing change writes swarm service labels only, which does
// not recreate the task: the process that locked everybody out is still running,
// and is what undoes it.
const (
	// The window is not the budget. Confirmation is refused for the first
	// entity.SettingsProbationSettleDelay of it, so what the operator actually has
	// is the window minus that - which is why the numbers here are not the round
	// ones somebody reading "you have two minutes" would expect.
	//
	// Erring long is the cheap direction. A window that runs out under a working
	// configuration throws away a change somebody had only to click to keep, and
	// that is the failure that makes people want the whole mechanism turned off. A
	// window that is too long costs the operator a longer wait to get back into a
	// dashboard - and only the dashboard: these settings govern the HivePaaS
	// routers alone, so no application traffic is riding on it.
	probationWindowDefault = 3 * time.Minute // 2m35s to answer
	probationWindowMin     = time.Minute     // 35s to answer
	probationWindowMax     = 15 * time.Minute

	// probationFallbackLag holds the in-process timer back so the queue, which is
	// the path with retries and a record, normally gets there first.
	probationFallbackLag = 20 * time.Second

	probationMaxRetry   = 3
	probationRetryDelay = 15 * time.Second
	probationTimeout    = 3 * time.Minute
)

// resolveProbationWindow clamps the requested window.
//
// There is deliberately no way to ask for zero. A caller that cannot confirm is
// exactly the caller this exists for: a script that applies a change and dies
// leaves the change reverted, which is the outcome we want and not one the
// script gets to opt out of.
func resolveProbationWindow(requested time.Duration) time.Duration {
	if requested <= 0 {
		return probationWindowDefault
	}
	return gofn.Clamp(requested, probationWindowMin, probationWindowMax)
}

// armProbation records the change as being on trial and schedules its undo.
//
// The task is built here but not scheduled: it goes into the same transaction as
// the change, so the deadline is committed with the thing it guards. Handing it
// to the queue is scheduleProbation's job, after the commit.
func (uc *UC) armProbation(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	data *updateRoutingSettingsData,
	persistingData *persistingAppData,
) error {
	timeNow := timeutil.NowUTC()
	setting := data.RoutingSetting

	snapshot := data.Snapshot
	pending, err := uc.findPendingProbation(ctx, db, data.App.ID)
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
		persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, pending)
		data.SupersededProbation = pending
	}

	deadlineAt := timeNow.Add(data.ProbationWindow)
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: data.App.ID,
		Type:     base.TaskTypeSettingsRevert,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityCritical,
			MaxRetry:   probationMaxRetry,
			RetryDelay: timeutil.Duration(probationRetryDelay),
			Timeout:    timeutil.Duration(probationTimeout),
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     deadlineAt,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	err = task.SetArgs(&entity.TaskSettingsRevertArgs{
		AppID:        data.App.ID,
		SettingID:    setting.ID,
		SettingType:  setting.Type,
		ProbationVer: setting.UpdateVer,
		Snapshot:     snapshot,
		AppliedBy:    authUserID(auth),
		AppliedAt:    timeNow,
		DeadlineAt:   deadlineAt,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, task)
	data.Probation = task
	return nil
}

// scheduleProbation hands the committed deadline to everything that can enforce it.
//
// Three triggers, one execution. The queue is the normal one; the in-process
// timer covers a worker that is not running at all, which is the default shape
// once RunWorkerInMainApp is turned off; the startup reconciler covers a deadline
// that passed while the process was down. They are safe to overlap because the
// task row is claimed with FOR UPDATE SKIP LOCKED and only in the not-started
// state, so whoever gets there first is the only one that acts.
//
// Nothing here is allowed to fail the change: by this point the row is committed,
// so the deadline exists whether or not anything managed to schedule it. A failure
// costs lateness, not the guarantee.
func (uc *UC) scheduleProbation(ctx context.Context, data *updateRoutingSettingsData) {
	if data == nil {
		return
	}
	if data.SupersededProbation != nil {
		if err := uc.taskQueue.UnscheduleTask(ctx, data.SupersededProbation); err != nil {
			uc.logger.Warnf("failed to unschedule superseded routing probation %s: %v",
				data.SupersededProbation.ID, err)
		}
	}
	if data.Probation == nil {
		return
	}
	if err := uc.taskQueue.ScheduleTask(ctx, data.Probation); err != nil {
		uc.logger.Warnf("failed to schedule routing probation %s, falling back to the local timer "+
			"and the startup scan: %v", data.Probation.ID, err)
	}
	uc.armProbationFallback(data.Probation) //nolint:contextcheck // the timer outlives this request
}

// armProbationFallback runs the revert from this process if nothing else did.
//
// The main app is the one place the revert is certain to be able to run: a
// routing change rewrites service labels, which swarm applies without recreating
// the task, so the process that published the change is still alive to undo it.
// The worker may be a separate service with zero replicas.
func (uc *UC) armProbationFallback(task *entity.Task) {
	delay := max(time.Until(task.RunAt)+probationFallbackLag, 0)
	taskID := task.ID
	time.AfterFunc(delay, func() {
		defer safego.RecoverWithLogger(uc.logger, "hpappsettingsuc.probationFallback")
		if err := uc.runProbationFallback(taskID); err != nil {
			uc.logger.Errorf("routing probation fallback failed for task %s: %v", taskID, err)
		}
	})
}

// runProbationFallback claims the task and reverts, if nothing else got there.
//
// The context is a fresh background one on purpose: the request that armed this
// returned minutes ago, and its context is long canceled. Inheriting it would
// cancel the revert exactly when it is most needed.
//
//nolint:contextcheck // deliberately detached from the request that armed it
func (uc *UC) runProbationFallback(taskID string) error {
	ctx := context.Background()
	return hperrors.Wrap(transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		task, err := uc.taskRepo.GetByID(ctx, db, base.TaskTypeSettingsRevert, taskID,
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
		return uc.executeProbationTask(ctx, db, task)
	}))
}

// executeProbationTask runs the revert and closes the task out.
//
// The caller must already hold the task row. Used by the local fallback and by
// the startup reconciler; the queue reaches the same work through the task
// executor, which does its own claiming.
func (uc *UC) executeProbationTask(ctx context.Context, db database.Tx, task *entity.Task) error {
	args, err := task.ArgsAsSettingsRevert()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("routing probation task %s has no args", task.ID)
	}

	resp, err := tasksettingsrevert.Run(ctx, db, uc.appRoutingService, uc.hpAppService, args)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if resp.Reverted {
		uc.logger.Warnf("reverted unconfirmed routing change on app %s, applied at %v",
			args.AppID, args.AppliedAt)
	}

	timeNow := timeutil.NowUTC()
	task.Status = base.TaskStatusDone
	task.StartedAt = timeNow
	task.EndedAt = timeNow
	task.UpdatedAt = timeNow
	task.MustSetOutput(&entity.TaskSettingsRevertOutput{Reverted: resp.Reverted, Reason: resp.Reason})
	return hperrors.Wrap(uc.taskRepo.Update(ctx, db, task))
}

// findPendingProbation returns the change currently on trial for the app.
//
// No row lock is taken here. Every writer reaches this after locking the routing
// setting itself, so that row is what serializes them; locking the task as well
// would add a second order to acquire and nothing else.
func (uc *UC) findPendingProbation(
	ctx context.Context,
	db database.IDB,
	appID string,
) (*entity.Task, error) {
	tasks, _, err := uc.taskRepo.List(ctx, db, "", nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSettingsRevert),
		bunex.SelectWhere("task.object_id = ?", appID),
		bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var latest *entity.Task
	for _, task := range tasks {
		if latest == nil || task.CreatedAt.After(latest.CreatedAt) {
			latest = task
		}
	}
	return latest, nil
}

func authUserID(auth *basedto.Auth) string {
	if auth == nil || auth.User == nil {
		return ""
	}
	return auth.User.ID
}

func (uc *UC) recordProbationOutcome(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	typ base.AuditLogType,
	setting *entity.Setting,
) error {
	return hperrors.Wrap(uc.auditService.Record(ctx, db, &auditservice.Entry{
		Type:    typ,
		Source:  base.AuditLogSourceAPIUpdate,
		Result:  base.AuditLogResultAllowed,
		Auth:    auth,
		ResType: base.ResourceTypeSetting,
		ResID:   setting.ID,
		ResName: setting.Name,
	}))
}

// ReconcileProbations picks up trials that nothing was left to enforce.
//
// The queue's own scan only looks in a window of TaskCheckInterval around now, so
// a deadline that passed while the process was down for longer than that is never
// rediscovered: the task stays not-started for good, and the change it was
// guarding becomes permanent. This is the sweep that closes that, and it also
// re-arms the local timer for trials still in their window.
//
// Called at startup, by the process that serves the API - the one a routing
// change cannot take down, because the change only rewrites service labels.
func (uc *UC) ReconcileProbations(ctx context.Context) error {
	tasks, _, err := uc.taskRepo.List(ctx, uc.db, "", nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSettingsRevert),
		bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	for _, task := range tasks {
		if task.RunAt.After(timeNow) {
			uc.armProbationFallback(task) //nolint:contextcheck // the timer outlives this call
			continue
		}
		uc.logger.Warnf("routing probation %s ran out at %v with nothing to enforce it, reverting now",
			task.ID, task.RunAt)
		if err := uc.runProbationFallback(task.ID); err != nil { //nolint:contextcheck // detached on purpose
			// One overdue trial failing is not a reason to abandon the rest, and
			// definitely not a reason to fail startup.
			uc.logger.Errorf("failed to revert overdue routing probation %s: %v", task.ID, err)
		}
	}
	return nil
}
