package entity

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type TaskDummyArgs struct {
	Sleep timeutil.Duration `json:"sleep"`
}

type TaskDummyOutput struct {
}

func (t *Task) ArgsAsDummy() (*TaskDummyArgs, error) {
	return parseTaskArgsAs(t, func() *TaskDummyArgs { return &TaskDummyArgs{} })
}

func (t *Task) OutputAsDummy() (*TaskDummyOutput, error) {
	return parseTaskOutputAs(t, func() *TaskDummyOutput { return &TaskDummyOutput{} })
}
