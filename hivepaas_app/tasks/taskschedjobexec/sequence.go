package taskschedjobexec

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

// executeSequence runs one step of a job sequence's run and says what comes
// next. The run is the task's output, saved after each step, and the task runs
// again for the next one (TaskExecData.Continue): a worker that restarts
// resumes at the step it was on, which then runs again.
//
// The first execution only prepares the run: it refuses to overlap another,
// and gives the task the first step's retry and timeout, which the queue
// reads before an execution starts.
func (e *Executor) executeSequence(ctx context.Context, db database.Tx, data *taskData) error {
	task := data.Task
	seqJob := data.SchedJob.MustAsSchedJob()
	run, err := task.OutputAsSchedJobSeqRun()
	if err != nil {
		return hperrors.Wrap(err)
	}

	if run == nil || !run.Started {
		return e.startSequence(ctx, db, data, seqJob)
	}
	if run.CurrentStep >= len(run.Steps) {
		return nil
	}

	step := run.Steps[run.CurrentStep]
	stepErr := e.runStep(ctx, db, data, run, step)
	if data.IsTaskCanceled() {
		markCanceled(run)
		e.logSummary(ctx, data, run)
		task.MustSetOutput(run)
		return nil
	}
	if stepErr != nil && retriesLeft(task.Config) {
		// The queue runs this step again after its retry delay.
		task.MustSetOutput(run)
		return hperrors.Wrap(stepErr)
	}
	if stepErr != nil {
		step.Status = base.SchedJobSeqStepFailed
	}

	next := afterStep(run, seqJob.Sequence.StopsOnFailure())
	switch next {
	case seqContinue:
		if err := e.prepareStep(ctx, db, task, seqJob, run.Steps[run.CurrentStep]); err != nil {
			return hperrors.Wrap(err)
		}
		task.MustSetOutput(run)
		data.Continue()
		return nil
	case seqFailed:
		e.logSummary(ctx, data, run)
		task.MustSetOutput(run)
		data.TaskNonRetryable = true
		return hperrors.NewArgumentInvalid("Job sequence").WithExtraDetail("%s", sequenceSummary(run))
	case seqDone:
		e.logSummary(ctx, data, run)
		task.MustSetOutput(run)
		return nil
	}
	return nil
}

// startSequence begins a run, unless another run of the sequence is still
// going: then this one ends at once, done, and says so.
func (e *Executor) startSequence(
	ctx context.Context,
	db database.Tx,
	data *taskData,
	seqJob *entity.SchedJob,
) error {
	task := data.Task
	busy, err := e.otherRunGoing(ctx, db, task)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if busy {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Skipped: the previous run of this sequence is still going.\n", tasklog.TsNow))
		data.SkipResultNotification = true
		return nil
	}

	run := entity.NewSchedJobSeqRun(seqJob.Sequence)
	run.Started = true
	if len(run.Steps) == 0 {
		task.MustSetOutput(run)
		return nil
	}
	if err := e.prepareStep(ctx, db, task, seqJob, run.Steps[0]); err != nil {
		return hperrors.Wrap(err)
	}
	task.MustSetOutput(run)
	data.Continue()
	return nil
}

// prepareStep gives the task the retry and timeout of the step's job, for the
// executions that run it. A job gone keeps the sequence's own.
func (e *Executor) prepareStep(
	ctx context.Context,
	db database.IDB,
	task *entity.Task,
	seqJob *entity.SchedJob,
	step *entity.SchedJobSeqStepResult,
) error {
	member, err := e.loadMember(ctx, db, step.Job.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	memberJob := &entity.SchedJob{}
	if member != nil {
		if memberJob, err = member.AsSchedJob(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	task.Config = stepConfig(task.Config, memberJob, seqJob)
	return nil
}

// runStep runs the step's job, and records how it went on the step, but for a
// failure, which the caller records once it knows no retry is left.
func (e *Executor) runStep(
	ctx context.Context,
	db database.Tx,
	data *taskData,
	run *entity.SchedJobSeqRun,
	step *entity.SchedJobSeqStepResult,
) error {
	logStore := data.LogStore
	member, err := e.loadMember(ctx, db, step.Job.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if member != nil && step.Name == "" {
		step.Name = member.Name
	}
	_ = logStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("\n── Step %d/%d: %s ──\n",
		run.CurrentStep+1, len(run.Steps), stepLabel(step)), tasklog.TsNow))

	if reason := memberSkipReason(member); reason != "" {
		step.Status = base.SchedJobSeqStepSkipped
		step.Error = reason
		_ = logStore.Add(ctx, tasklog.NewWarnFrame("Skipped: "+reason+"\n", tasklog.TsNow))
		return nil
	}

	step.Status = base.SchedJobSeqStepRunning
	step.Attempts = data.Task.Config.Retry + 1
	step.StartedAt = timeutil.NowUTC()
	step.Error = ""

	refObjects, err := e.loadMemberRefObjects(ctx, db, member)
	if err != nil {
		step.Error = stepError(err)
		return hperrors.Wrap(err)
	}
	result, err := e.runJob(ctx, db, &jobRun{
		execData:   data.SubTask(data.Task),
		jobSetting: member,
		refObjects: refObjects,
		sequence: &schedjobexecservice.SequenceStep{
			Step:       run.CurrentStep + 1,
			Steps:      len(run.Steps),
			Earlier:    run.Steps[:run.CurrentStep],
			OutputFile: schedjobexecservice.OutputFilePath(data.Task.ID, run.CurrentStep+1),
		},
	})
	step.EndedAt = timeutil.NowUTC()
	if result != nil {
		step.ExitCode = result.exitCode
		step.Outputs = result.outputs
	}
	if err != nil {
		step.Error = stepError(err)
		return hperrors.Wrap(err)
	}
	step.Status = base.SchedJobSeqStepDone
	return nil
}

// memberSkipReason says why a step's job is not run, or nothing when it is.
func memberSkipReason(member *entity.Setting) string {
	switch {
	case member == nil:
		return "the scheduled job is gone"
	case !member.IsActive():
		return "the scheduled job is disabled"
	case member.Kind == string(base.SchedJobTypeJobSequence):
		return "a sequence cannot run another sequence"
	}
	return ""
}

// loadMember is a step's scheduled job, whatever its scope and status; nil
// when it is gone.
func (e *Executor) loadMember(ctx context.Context, db database.IDB, jobID string) (*entity.Setting, error) {
	refObjects := entity.NewRefObjects()
	err := e.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects, nil, false,
		&entity.RefObjectIDs{RefSettingIDs: []string{jobID}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	member := refObjects.RefSettings[jobID]
	if member != nil && member.Type != base.SettingTypeSchedJob {
		return nil, nil //nolint:nilnil // not a scheduled job: as good as gone
	}
	return member, nil
}

// loadMemberRefObjects is what the job references, in its own scope, as when
// it runs on its own schedule.
func (e *Executor) loadMemberRefObjects(
	ctx context.Context,
	db database.IDB,
	member *entity.Setting,
) (*entity.RefObjects, error) {
	scope, err := e.scopeService.LoadObjectScope(ctx, db, member.Scope, member.ObjectID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refObjects := entity.NewRefObjects()
	refObjects.AddObjectScope(scope)
	if err = e.settingService.LoadRefObjectsSkipMissing(ctx, db, &refObjects, scope, true, member); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return refObjects, nil
}

// otherRunGoing says whether another run of the task's sequence has begun and
// not ended: started, and neither done, canceled nor failed for good.
func (e *Executor) otherRunGoing(ctx context.Context, db database.IDB, task *entity.Task) (bool, error) {
	tasks, _, err := e.taskRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("task.target_id = ?", task.TargetID),
		bunex.SelectWhere("task.type = ?", base.TaskTypeSchedJobExec),
		bunex.SelectWhere("task.id != ?", task.ID),
		bunex.SelectWhere("task.started_at IS NOT NULL"),
		bunex.SelectWhereGroup(
			bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
			bunex.SelectWhereOrGroup(
				bunex.SelectWhere("task.status = ?", base.TaskStatusFailed),
				bunex.SelectWhere("task.retry_at IS NOT NULL"),
			),
		),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(tasks) > 0, nil
}

func (e *Executor) logSummary(ctx context.Context, data *taskData, run *entity.SchedJobSeqRun) {
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("\n── Summary ──\n"+sequenceSummary(run)+"\n", tasklog.TsNow))
}

// stepError is how a step's failure reads in its result: the error's detail,
// else its text.
func stepError(err error) string {
	if detail := hperrors.GetErrorDetail(err, ""); detail != "" {
		return detail
	}
	return err.Error()
}
