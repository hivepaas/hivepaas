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

	// ListSnapshots reads the snapshots currently stored in a backup repository.
	ListSnapshots(ctx context.Context, db database.IDB, req *ListSnapshotsReq) (
		resp *ListSnapshotsResp, err error)
}
