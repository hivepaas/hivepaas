package taskmanagedrestore

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

const (
	// waitSlice is how long one run waits on a step before it continues the
	// task: a recovery takes as long as the WAL it replays, hours for a large
	// one, and no run holds its transaction that long.
	waitSlice = 3 * time.Minute
	pollEvery = 10 * time.Second
	// recoveryDeadline is how long a recovery may take before the restore gives
	// up on it.
	recoveryDeadline = 48 * time.Hour
	// serviceUpdateRetries retries a service update that met another.
	serviceUpdateRetries = 2
)

// waitRecovery asks the app whether its recovery has ended, until it has, it
// failed, or this run's slice is up. The app not running yet - its first
// deployment still going - is waited for as a recovery is.
func (e *Executor) waitRecovery(ctx context.Context, db database.Tx, run *restoreRun) error {
	now := e.now()
	if run.state.RecoverySince.IsZero() {
		run.state.RecoverySince = now
		e.log(ctx, run, "Waiting for the database to recover...")
	}
	if now.Sub(run.state.RecoverySince) > recoveryDeadline {
		return failed(run, fmt.Sprintf("the recovery has not ended in %s", recoveryDeadline))
	}
	check := run.engine.RecoveryCheck()
	until := now.Add(waitSlice)
	for {
		resp, _ := e.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
			TaskExecData: run.data, App: run.app, Command: &entity.CommandTemplate{Command: check.Command}})
		if resp != nil && resp.ExitCode != nil {
			switch *resp.ExitCode {
			case 0:
				e.log(ctx, run, "Recovered: the server is up, and promoted.")
				run.state.Step = entity.ManagedRestoreStepBackup
				run.data.Continue()
				return nil
			case managedmodel.RecoveryExitRecovering:
			default:
				return failed(run, "the recovery failed, as the database said above")
			}
		}
		if !e.now().Before(until) {
			run.data.Continue()
			return nil
		}
		if err := e.sleep(ctx, pollEvery); err != nil {
			return err
		}
	}
}

// startBackup starts the engine's jobs, which waited for the recovery, and runs
// the app's nightly job now: its first base backup, which starts its own
// restore window, and which Status reads as any other run.
func (e *Executor) startBackup(ctx context.Context, db database.Tx, run *restoreRun) error {
	jobs, _, err := e.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", run.app.ID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSchedJob),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	now := e.now()
	var nightly *entity.Setting
	for _, setting := range jobs {
		job, jobErr := setting.AsSchedJob()
		if jobErr != nil || job.ManagedBy.ID != run.setting.ID || !setting.DeletedAt.IsZero() {
			continue
		}
		if setting.ID == run.engine.FirstBackupJob(run.managed) {
			nightly = setting
		}
		if setting.Status == base.SettingStatusActive {
			continue
		}
		setting.Status = base.SettingStatusActive
		setting.UpdateVer++
		setting.UpdatedAt = now
		if err = e.settingRepo.Update(ctx, db, setting, bunex.UpdateColumns("status", "update_ver",
			"updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
		if err = e.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, false); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if nightly == nil {
		return failed(run, "the job that takes its base backups is gone")
	}

	task, err := e.schedJobService.CreateSchedJobTask(nightly, now, now)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = e.taskRepo.Insert(ctx, db, task); err != nil {
		return hperrors.Wrap(err)
	}
	// Once this run's transaction has committed: the task's row is in it.
	run.data.OnPostTx(func() { _ = e.taskQueue.ScheduleTask(context.Background(), task) }) //nolint:contextcheck
	e.log(ctx, run, fmt.Sprintf("Taking the first base backup: run %s of the job %q.", task.ID, nightly.Name))
	run.state.BackupTask = task.ID
	run.state.Step = entity.ManagedRestoreStepWaitBackup
	run.data.Continue()
	return nil
}

// waitBackup waits for the first base backup's run to end.
func (e *Executor) waitBackup(ctx context.Context, db database.Tx, run *restoreRun) error {
	until := e.now().Add(waitSlice)
	for {
		task, err := e.taskRepo.GetByID(ctx, db, nil, base.TaskTypeSchedJobExec, run.state.BackupTask)
		if errors.Is(err, hperrors.ErrNotFound) {
			return failed(run, "the first base backup's run is gone")
		}
		if err != nil {
			return hperrors.Wrap(err)
		}
		switch {
		case task.Status == base.TaskStatusDone:
			e.log(ctx, run, "The first base backup is taken.")
			run.state.Step = entity.ManagedRestoreStepFinish
			run.data.Continue()
			return nil
		case task.Status == base.TaskStatusCanceled || (task.Status == base.TaskStatusFailed && !task.CanRetry()):
			return failed(run, fmt.Sprintf("the first base backup failed: see run %s", task.ID))
		}
		if !e.now().Before(until) {
			run.data.Continue()
			return nil
		}
		if err = e.sleep(ctx, pollEvery); err != nil {
			return err
		}
	}
}

// finish takes off the app what it was restored with - the source's variables,
// its keys, its mount - and marks it ready.
func (e *Executor) finish(ctx context.Context, db database.Tx, run *restoreRun) error {
	settings, _, err := e.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", run.app.ID),
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeEnvVar, base.SettingTypeSecret,
			base.SettingTypeAppSettingMount),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	now := e.now()
	resp, err := run.engine.FinishRestore(&managedmodel.FinishRestoreReq{Managed: run.managed, Settings: settings,
		Now: now})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = e.persistFinish(ctx, db, run, resp, now); err != nil {
		return err
	}
	if err = e.applyFinish(ctx, db, run, resp); err != nil {
		return err
	}
	e.log(ctx, run, "Ready: the database keeps nothing of the source's archive, and archives to its own.")
	run.state.Step = entity.ManagedRestoreStepDone
	return nil
}

func (e *Executor) persistFinish(ctx context.Context, db database.Tx, run *restoreRun,
	resp *managedmodel.FinishRestoreResp, now time.Time) error {
	for _, setting := range resp.Changed {
		setting.UpdatedAt = now
		if err := e.settingRepo.Update(ctx, db, setting,
			bunex.UpdateColumns("data", "update_ver", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
	}
	for _, setting := range resp.Deleted {
		setting.DeletedAt = now
		setting.UpdateVer++
		if err := e.settingRepo.Update(ctx, db, setting, bunex.UpdateColumns("deleted_at", "update_ver")); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if err := run.setting.SetData(run.managed); err != nil {
		return hperrors.Wrap(err)
	}
	run.setting.UpdateVer++
	run.setting.UpdatedAt = now
	return hperrors.Wrap(e.settingRepo.Update(ctx, db, run.setting,
		bunex.UpdateColumns("data", "update_ver", "updated_at")))
}

// applyFinish brings the app's service in line: its variables built again, the
// source's mount taken off, and the keys' files unmounted by a refresh.
func (e *Executor) applyFinish(ctx context.Context, db database.Tx, run *restoreRun,
	resp *managedmodel.FinishRestoreResp) error {
	if len(resp.EnvKeys) > 0 {
		appEnv, err := e.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, run.app.GetObjectScope(), false,
			[]string{run.app.ID}, false, false)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if errs := e.envVarService.ApplyEnvVarsForApps(ctx, db, appEnv, false, false); len(errs) > 0 {
			return hperrors.Wrap(errors.Join(slices.Collect(maps.Values(errs))...))
		}
	}
	if len(resp.MountTargets) > 0 {
		err := e.dockerManager.ServiceUpdateFunc(ctx, run.app.ServiceID, nil,
			func(_ int, svc *swarm.Service) (bool, error) {
				spec := svc.Spec.TaskTemplate.ContainerSpec
				kept := slices.DeleteFunc(slices.Clone(spec.Mounts), func(m mount.Mount) bool {
					return slices.Contains(resp.MountTargets, m.Target)
				})
				if len(kept) == len(spec.Mounts) {
					return false, nil
				}
				spec.Mounts = kept
				return true, nil
			}, serviceUpdateRetries, 0)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	if len(resp.Deleted) > 0 {
		refresh, err := e.settingMountService.RecordRefresh(ctx, db, resp.Deleted...)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if refresh != nil {
			run.data.OnPostTx(func() { _ = e.taskQueue.ScheduleTask(context.Background(), refresh) }) //nolint:contextcheck
		}
	}
	return nil
}

// failed ends the restore with reason: what failed is not done again by itself.
func failed(run *restoreRun, reason string) error {
	run.data.TaskNonRetryable = true
	return hperrors.Wrap(hperrors.ErrAppManagedRestoreFailed).WithParam("Reason", reason).ExposeParams("Reason")
}

func (e *Executor) log(ctx context.Context, run *restoreRun, line string) {
	_ = run.data.LogStore.Add(ctx, tasklog.NewOutFrame(line+"\n", tasklog.TsNow))
}
