package backupreposervice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// InitRepo creates a new backup repository on the storage backend, or connects to an existing
	// one when ImportExisting is set. With SyncData it also reads back the snapshots already
	// present in the repository so the caller can persist them.
	InitRepo(ctx context.Context, db database.IDB, req *InitRepoReq) (resp *InitRepoResp, err error)
	CleanupRepo(ctx context.Context, db database.IDB, req *CleanupRepoReq) (resp *CleanupRepoResp, err error)

	// ListSnapshots reads the snapshots currently stored in a backup repository.
	ListSnapshots(ctx context.Context, db database.IDB, req *ListSnapshotsReq) (
		resp *ListSnapshotsResp, err error)

	// ChangeRepoPassword re-encrypts the repository with a new password. The repository itself is
	// the source of truth: on success the old password stops working immediately.
	ChangeRepoPassword(ctx context.Context, db database.IDB, req *ChangeRepoPasswordReq) error

	// ApplyRepoOptions pushes the changeable repository settings onto the repository. They are
	// stored inside it, so backups taken from any node pick them up without further work.
	ApplyRepoOptions(ctx context.Context, db database.IDB, req *ApplyRepoOptionsReq) error
}
