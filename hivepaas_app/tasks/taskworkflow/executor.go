package taskworkflow

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	taskQueue queue.TaskQueue
}

func NewExecutor(
	taskQueue queue.TaskQueue,
) *Executor {
	e := &Executor{
		taskQueue: taskQueue,
	}
	taskQueue.RegisterExecutor(base.TaskTypeWorkflow, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	execData *queue.TaskExecData,
) error {
	task := execData.Task
	wfArgs, err := task.ArgsAsWorkflow()
	if err != nil {
		return apperrors.Wrap(err)
	}

	if wfArgs == nil || len(wfArgs.Steps) == 0 {
		return nil
	}

	// All steps completed
	if wfArgs.CurrentStep >= len(wfArgs.Steps) {
		return nil
	}

	step := wfArgs.Steps[wfArgs.CurrentStep]

	targetID := step.TargetID
	if targetID == "" {
		targetID = task.TargetID
	}

	subTask := &entity.Task{
		ID:       fmt.Sprintf("%s-step-%d", task.ID, wfArgs.CurrentStep),
		Scope:    task.Scope,
		ObjectID: task.ObjectID,
		TargetID: targetID,
		Type:     step.Type,
		Config:   e.getStepConfig(task.Config, step.Config),
		Args:     step.Args,
	}

	subTaskExecData := &queue.TaskExecData{
		Task: subTask,
	}

	// Execute the current step
	err = e.taskQueue.ExecuteTaskType(ctx, db, step.Type, subTaskExecData)
	if err != nil {
		if !step.IgnoreError {
			return apperrors.Wrap(err).WithMsgLog("workflow %s failed at step %d (%s)",
				task.ID, wfArgs.CurrentStep, step.Name)
		}
	}

	// Advance to next step
	wfArgs.CurrentStep++
	if err := task.SetArgs(wfArgs); err != nil {
		return apperrors.Wrap(err)
	}

	// If there are remaining steps, re-enqueue workflow task for next step
	if wfArgs.CurrentStep < len(wfArgs.Steps) {
		execData.OnPostTransaction(func() { //nolint:contextcheck
			_ = e.taskQueue.ScheduleTask(context.Background(), task)
		})
	}

	return nil
}

func (e *Executor) getStepConfig(
	parentConfig entity.TaskConfig,
	stepConfig *entity.TaskConfig,
) entity.TaskConfig {
	finalConfig := parentConfig

	if stepConfig == nil {
		return finalConfig
	}

	if stepConfig.Priority != "" {
		finalConfig.Priority = stepConfig.Priority
	}
	if stepConfig.MaxRetry > 0 {
		finalConfig.MaxRetry = stepConfig.MaxRetry
	}
	if stepConfig.RetryDelay > 0 {
		finalConfig.RetryDelay = stepConfig.RetryDelay
	}
	if stepConfig.RetryDelayIncr > 0 {
		finalConfig.RetryDelayIncr = stepConfig.RetryDelayIncr
	}
	if stepConfig.RetryBackoffJitter > 0 {
		finalConfig.RetryBackoffJitter = stepConfig.RetryBackoffJitter
	}
	if stepConfig.RetryDelayMax > 0 {
		finalConfig.RetryDelayMax = stepConfig.RetryDelayMax
	}
	if stepConfig.Timeout > 0 {
		finalConfig.Timeout = stepConfig.Timeout
	}
	if stepConfig.ControlDisabled {
		finalConfig.ControlDisabled = stepConfig.ControlDisabled
	}

	return finalConfig
}
