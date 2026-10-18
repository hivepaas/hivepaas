package entity

import "time"

// TaskManagedRestoreArgs is a restore of a managed database into a new app, as
// it was asked for. The task's target and object are the new app.
type TaskManagedRestoreArgs struct {
	ProjectID   string `json:"projectId"`
	AppID       string `json:"appId"`
	SourceAppID string `json:"sourceAppId"`
}

func (t *Task) ArgsAsManagedRestore() (*TaskManagedRestoreArgs, error) {
	return parseTaskArgsAs(t, func() *TaskManagedRestoreArgs { return &TaskManagedRestoreArgs{} })
}

// Where a restore is: each step can run again, and resumes where it was.
const (
	ManagedRestoreStepRecover    = "recover"
	ManagedRestoreStepBackup     = "backup"
	ManagedRestoreStepWaitBackup = "wait-backup"
	ManagedRestoreStepFinish     = "finish"
	ManagedRestoreStepDone       = "done"
)

// TaskManagedRestoreState is the task's output: the step it is at, and what the
// steps before it left for it.
type TaskManagedRestoreState struct {
	Step string `json:"step"`
	// RecoverySince is when the restore began waiting for the recovery to end.
	RecoverySince time.Time `json:"recoverySince,omitzero"`
	// BackupTask is the run of the app's nightly job its first base backup is.
	BackupTask string `json:"backupTask,omitempty"`
}

func (t *Task) OutputAsManagedRestore() (*TaskManagedRestoreState, error) {
	return parseTaskOutputAs(t, func() *TaskManagedRestoreState { return &TaskManagedRestoreState{} })
}
