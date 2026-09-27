package backupreposervice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// InitRepo creates a new backup repository on the storage backend, or connects to an existing
	// one when ImportExisting is set. With SyncData it also reads back the snapshots already
	// present in the repository so the caller can persist them.
	InitRepo(ctx context.Context, db database.IDB, req *InitRepoReq) (resp *InitRepoResp, err error)
	CleanupRepo(ctx context.Context, db database.IDB, req *CleanupRepoReq) (resp *CleanupRepoResp, err error)

	// SyncRepo reads a repository back: the options it is configured with, and the snapshots it
	// holds. It writes nothing, on the repository or in the DB - it is what the caller needs to
	// adopt changes made outside the app.
	SyncRepo(ctx context.Context, db database.IDB, req *SyncRepoReq) (resp *SyncRepoResp, err error)

	// SyncRepoSnapshots reconciles the stored snapshot records against what the repository holds.
	// It runs on the caller's transaction.
	SyncRepoSnapshots(ctx context.Context, db database.Tx, req *SyncRepoSnapshotsReq) (
		resp *SyncRepoSnapshotsResp, err error)

	// ListSnapshots reads the snapshots currently stored in a backup repository.
	ListSnapshots(ctx context.Context, db database.IDB, req *ListSnapshotsReq) (
		resp *ListSnapshotsResp, err error)

	// ChangeRepoPassword re-encrypts the repository with a new password. The repository itself is
	// the source of truth: on success the old password stops working immediately.
	ChangeRepoPassword(ctx context.Context, db database.IDB, req *ChangeRepoPasswordReq) error

	// ApplyRepoOptions pushes the changeable repository settings onto the repository. They are
	// stored inside it, so backups taken from any node pick them up without further work.
	ApplyRepoOptions(ctx context.Context, db database.IDB, req *ApplyRepoOptionsReq) error

	// BackupStream takes a snapshot of what Stdin carries, as the file FileName, the engine
	// running in this process. A repository on a volume cannot take one yet: the agent that
	// reaches it carries no stdin.
	BackupStream(ctx context.Context, db database.IDB, req *BackupStreamReq) (*BackupResp, error)
	// BackupDirectory takes a snapshot of a directory of a node's host, the engine running on
	// that node through its agent. A repository on a volume is reachable on its own node only.
	// OpenRepoServer runs a repository server for a repository on a volume, on
	// its node, until the session is closed or ctx ends.
	OpenRepoServer(ctx context.Context, db database.IDB, req *OpenRepoServerReq) (*RepoServerSession, error)
	BackupDirectory(ctx context.Context, db database.IDB, req *BackupDirectoryReq) (*BackupResp, error)
	// DeleteSnapshot removes one snapshot from the repository.
	DeleteSnapshot(ctx context.Context, db database.IDB, req *DeleteSnapshotReq) error
	// VolumeHostDir is where a volume's data is on the host of its node, and that node.
	VolumeHostDir(ctx context.Context, volume *entity.Setting) (*VolumeHostDir, error)
}
