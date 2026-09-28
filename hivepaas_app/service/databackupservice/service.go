package databackupservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// Service runs data-backup jobs: a snapshot of an app's data into a backup repository.
type Service interface {
	// Backup takes the snapshot a data-backup job says, records it in the repository's
	// snapshot list, and, but for a step of a job sequence, in the task's output.
	Backup(ctx context.Context, db database.Tx, req *BackupReq) (*BackupResp, error)
	// Restore puts a snapshot back into an app: a command snapshot through a
	// command run in the app, a volume snapshot into the app's directory of a
	// volume. It runs on the task's transaction.
	Restore(ctx context.Context, db database.Tx, req *RestoreReq) error
	// CheckAppVolume says whether the app mounts its own directory of the volume:
	// what a volume source reads.
	CheckAppVolume(ctx context.Context, db database.IDB, app *entity.App, volumeID string) error
	// FindAppVolume is where the app's part of a volume it mounts is on a host; an error
	// when the app does not mount it as its own directory.
	FindAppVolume(ctx context.Context, db database.IDB, app *entity.App, volumeID string) (*AppVolume, error)
}

type BackupReq struct {
	*queue.TaskExecData
	JobSetting *entity.Setting
	App        *entity.App
	// RefObjects holds what the job references: its repository and its volume.
	RefObjects *entity.RefObjects
	// Sequence is set when the job runs as a step of a job sequence.
	Sequence *schedjobexecservice.SequenceStep
}

type BackupResp struct {
	Result *entity.SchedJobDataBackupResult
}

// AppVolume is the app's part of a volume, on the host of the volume's node.
type AppVolume struct {
	HostDir   string
	NodeID    string
	NodeLabel string
}

type RestoreReq struct {
	*queue.TaskExecData
	// Target is the repository the snapshot is in.
	Target     backupreposervice.RepoTarget
	SnapshotID string
	// App is the app the data goes into.
	App *entity.App
	// One of these, as the snapshot is.
	Command *RestoreCommand
	Volume  *RestoreVolume
}

// RestoreCommand is a command snapshot's restore: the command that loads the
// snapshot's file from its stdin.
type RestoreCommand struct {
	Command  *entity.CommandTemplate
	FileName string
}

// RestoreVolume is a volume snapshot's restore.
type RestoreVolume struct {
	// VolumeID is a volume the app mounts as its own directory, and Subpath a
	// path inside what the app sees of it: where the snapshot's root goes.
	VolumeID string
	Subpath  string
	// SnapshotPath is a directory inside the snapshot to restore alone, into the
	// same place under Subpath; "" for all of it.
	SnapshotPath string
	StopApp      bool
	Mode         base.BackupRestoreMode
}
