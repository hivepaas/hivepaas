package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

type TaskSystemUpdateArgs struct {
	CurrentVersion *base.ReleaseInfo `json:"currentVersion"`
	TargetVersion  *base.ReleaseInfo `json:"targetVersion"`

	// SkipBackup runs the update without dumping the database first.
	//
	// The dump is what a failed migration is undone from, so this is asking to
	// update without a way back. It exists because the alternative is worse: an
	// installation whose disk is full, or whose pg_dump is missing, would
	// otherwise be unable to update at all.
	SkipBackup bool `json:"skipBackup,omitempty"`
}

type TaskSystemUpdateOutput struct {
}

func (t *Task) ArgsAsSystemUpdate() (*TaskSystemUpdateArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSystemUpdateArgs { return &TaskSystemUpdateArgs{} })
}

func (t *Task) OutputAsSystemUpdate() (*TaskSystemUpdateOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSystemUpdateOutput { return &TaskSystemUpdateOutput{} })
}
