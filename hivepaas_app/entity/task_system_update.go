package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

type TaskSystemUpdateArgs struct {
	CurrentVersion *base.ReleaseInfo `json:"currentVersion"`
	TargetVersion  *base.ReleaseInfo `json:"targetVersion"`
}

type TaskSystemUpdateOutput struct {
}

func (t *Task) ArgsAsSystemUpdate() (*TaskSystemUpdateArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSystemUpdateArgs { return &TaskSystemUpdateArgs{} })
}

func (t *Task) OutputAsSystemUpdate() (*TaskSystemUpdateOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSystemUpdateOutput { return &TaskSystemUpdateOutput{} })
}
