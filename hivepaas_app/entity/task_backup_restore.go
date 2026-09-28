package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// TaskBackupRestoreArgs is a restore, as it was asked for: which snapshot, into
// which app, and how. The task's target is the snapshot's record, its object the
// app.
type TaskBackupRestoreArgs struct {
	ProjectID string `json:"projectId"`
	AppID     string `json:"appId"`
	RepoID    string `json:"repoId"`
	// SnapshotID is the repository's, in full.
	SnapshotID string `json:"snapshotId"`

	// A command snapshot: the command its file is loaded with.
	Command  *CommandTemplate `json:"command,omitempty"`
	FileName string           `json:"fileName,omitempty"`

	// A volume snapshot: where it goes, what of it, and how.
	Volume       ObjectID               `json:"volume,omitzero"`
	Subpath      string                 `json:"subpath,omitempty"`
	SnapshotPath string                 `json:"snapshotPath,omitempty"`
	StopApp      bool                   `json:"stopApp,omitempty"`
	Mode         base.BackupRestoreMode `json:"mode,omitempty"`
}

func (t *Task) ArgsAsBackupRestore() (*TaskBackupRestoreArgs, error) {
	return parseTaskArgsAs(t, func() *TaskBackupRestoreArgs { return &TaskBackupRestoreArgs{} })
}
