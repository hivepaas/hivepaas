package backupreposervice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/backup"
)

type InitRepoReq struct {
	Scope *entity.ObjectScope
	Repo  *entity.BackupRepo
	// RepoID is the ID of the setting holding the repo. It keeps the engine connection state of
	// this repo isolated from every other repo. Optional: a temporary ID is derived when empty.
	RepoID         string
	RepoName       string
	ImportExisting bool
	SyncData       bool
	RefObjects     *entity.RefObjects
}

type InitRepoResp struct {
	// Snapshots are the snapshots found in an imported repository. Always empty when creating a
	// new repository, and when SyncData is off.
	Snapshots []*RepoSnapshot
}

// RepoSnapshot pairs a snapshot with its tags. Tags are stored in the tags table rather than
// inside the snapshot data so they can be indexed and searched, so they travel alongside it.
type RepoSnapshot struct {
	Snapshot *entity.BackupSnapshot
	Tags     []string
}

type ListSnapshotsReq struct {
	Scope      *entity.ObjectScope
	Repo       *entity.BackupRepo
	RepoID     string
	RefObjects *entity.RefObjects
	Options    *backup.ListSnapshotsOptions
}

type ListSnapshotsResp struct {
	Snapshots []*RepoSnapshot
}
