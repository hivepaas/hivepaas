package taskschedjobexec

import (
	"context"
	"maps"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

// runNightly runs a managed database's nightly job in two steps: its base
// backup, then the retention its engine decides from the backups the first step
// listed and the restore points kept (design §4, "Retention and kept points").
// A backup that failed is not followed by retention. The run's outputs are both
// steps', the retention's over the backup's. The backup is taken in the app the
// job names, which may be a replica, sparing the database; retention and the
// pruning after it run on the database the job is the engine's of.
func (e *Executor) runNightly(
	ctx context.Context,
	db database.Tx,
	run *jobRun,
	job *entity.SchedJob,
) (*jobResult, error) {
	managed, engine, err := e.managedEngineOf(run, job)
	if err != nil {
		return nil, err
	}
	primary := run.refObjects.RefApps[run.jobSetting.ObjectID]
	backupOn := run.refObjects.RefApps[job.App.ID]
	if primary == nil {
		primary = backupOn
	}
	if backupOn == nil {
		// The replica named is gone: the database takes its own backup.
		_ = run.execData.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"The replica the base backups are taken on is gone: this one is taken on the database.\n",
			tasklog.TsNow))
		backupOn = primary
	}
	result := &jobResult{outputs: map[string]string{}}

	resp, err := e.execNightlyStep(ctx, db, run, backupOn, job, engine.NightlyBackupArgs())
	result.add(resp)
	if err != nil {
		return result, hperrors.Wrap(err)
	}
	retention, err := engine.Retention(&managedmodel.RetentionReq{Managed: managed, Outputs: result.outputs,
		Now: timeutil.NowUTC()})
	if err != nil {
		return result, hperrors.Wrap(err)
	}
	line := "Retention deletes nothing: there are not more base backups than it keeps.\n"
	if retention.Boundary != "" {
		line = "Retention keeps " + retention.Boundary + " and every base backup after it.\n"
	}
	_ = run.execData.LogStore.Add(ctx, tasklog.NewOutFrame(line, tasklog.TsNow))

	resp, err = e.execNightlyStep(ctx, db, run, primary, job, retention.Args)
	result.add(resp)
	if err != nil {
		return result, hperrors.Wrap(err)
	}
	e.pruneDeployPoints(ctx, db, run, primary, result.outputs)
	return result, nil
}

// pruneDeployPoints removes the points deployments made that retention left no
// base backup for, from what the retention step listed. It never fails the
// run: its backups are made and kept either way.
func (e *Executor) pruneDeployPoints(ctx context.Context, db database.Tx, run *jobRun, app *entity.App,
	outputs map[string]string) {
	pruned, err := e.managedService.PruneDeployPoints(ctx, db, app.ProjectID, app.ID, outputs)
	switch {
	case err != nil:
		_ = run.execData.LogStore.Add(ctx, tasklog.NewErrFrame(
			"The deployments' restore points were not checked: "+hperrors.GetErrorDetail(err, "")+"\n",
			tasklog.TsNow))
	case len(pruned) > 0:
		_ = run.execData.LogStore.Add(ctx, tasklog.NewOutFrame(
			"No base backup kept reaches these deployments' restore points any more; they are removed: "+
				strings.Join(pruned, ", ")+"\n", tasklog.TsNow))
	}
}

func (e *Executor) execNightlyStep(ctx context.Context, db database.Tx, run *jobRun, app *entity.App,
	job *entity.SchedJob, args string) (*schedjobexecservice.SchedJobExecResp, error) {
	command := *job.Command
	command.Command += " " + args
	resp, err := e.schedJobExecService.SchedJobExec(ctx, db, &schedjobexecservice.SchedJobExecReq{
		TaskExecData:    run.execData,
		SchedJobSetting: run.jobSetting,
		DestApp:         app,
		Sequence:        run.sequence,
		Command:         &command,
	})
	return resp, hperrors.Wrap(err)
}

// add takes a step's exit code and outputs into the run's.
func (r *jobResult) add(resp *schedjobexecservice.SchedJobExecResp) {
	if resp == nil {
		return
	}
	r.exitCode = resp.ExitCode
	r.skipNotification = resp.SkipResultNotification
	maps.Copy(r.outputs, resp.Outputs)
}

// managedEngineOf is the engine's setting a managed job belongs to, and the
// engine.
func (e *Executor) managedEngineOf(run *jobRun, job *entity.SchedJob) (
	*entity.AppManagedSettings, managedmodel.Engine, error) {
	setting := run.refObjects.RefSettings[job.ManagedBy.ID]
	if setting == nil || job.Command == nil {
		return nil, nil, hperrors.NewNotFound("Managed engine")
	}
	managed, err := setting.AsAppManagedSettings()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	engine, found := e.managedService.Engine(templatemodel.ManagedEngine(managed.Engine))
	if !found {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).
			WithExtraDetail("engine %s is not one this HivePaaS has", managed.Engine)
	}
	return managed, engine, nil
}
