package databackupservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// Service runs data-backup jobs: a snapshot of an app's data into a backup repository.
type Service interface {
	// Backup takes the snapshot a data-backup job says, records it in the repository's
	// snapshot list, and, but for a step of a job sequence, in the task's output.
	Backup(ctx context.Context, db database.Tx, req *BackupReq) (*BackupResp, error)
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
