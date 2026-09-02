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

	// Config is what an imported repository is actually configured with. Nil when creating a new
	// repository, where the request is what decides the settings.
	Config *backup.RepoConfig
}

type CleanupRepoReq struct {
	Scope       *entity.ObjectScope
	RepoSetting *entity.Setting
	RefObjects  *entity.RefObjects
}

type CleanupRepoResp struct {
	// Remaining is what the repository holds after the prune. The engine reports only a per-source
	// count of what it expired - no IDs, no JSON - so this list is what the caller has to
	// reconcile its own records against. Listing the repository beforehand would add nothing: the
	// stored records already say what the app believed was there.
	Remaining []*RepoSnapshot
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

type ChangeRepoPasswordReq struct {
	Scope      *entity.ObjectScope
	Repo       *entity.BackupRepo
	RepoID     string
	RefObjects *entity.RefObjects

	// OldPassword is what the repository is encrypted with right now. It overrides the password
	// held by Repo, which lets a failed change be rolled back by swapping the two.
	OldPassword string
	NewPassword string
}

type ApplyRepoOptionsReq struct {
	Scope      *entity.ObjectScope
	Repo       *entity.BackupRepo
	RepoID     string
	RefObjects *entity.RefObjects
	Options    *backup.RepoOptions
}
