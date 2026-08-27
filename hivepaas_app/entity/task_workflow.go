package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type WorkflowStep struct {
	Name        string        `json:"name"`
	Type        base.TaskType `json:"type"`
	Args        string        `json:"args"`
	TargetID    string        `json:"targetId,omitempty"`
	Config      *TaskConfig   `json:"config,omitempty"`
	IgnoreError bool          `json:"ignoreError,omitempty"`
}

type WorkflowArgs struct {
	CurrentStep int             `json:"currentStep"`
	Steps       []*WorkflowStep `json:"steps"`
}

func (t *Task) ArgsAsWorkflow() (*WorkflowArgs, error) {
	if t.Type != base.TaskTypeWorkflow {
		return nil, hperrors.NewMismatch("Task type", base.TaskTypeWorkflow)
	}
	return parseTaskArgsAs(t, func() *WorkflowArgs {
		return &WorkflowArgs{}
	})
}
