package backupreposervice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
	"github.com/hivepaas/hivepaas/services/backup"
)

var (
	PasswordRequirements = secrethelper.SecretStrengthRequirements{
		MinLen:             secrethelper.DefaultSecretMinLen,
		MaxLen:             secrethelper.DefaultSecretMaxLen,
		RequiredLowercases: secrethelper.DefaultSecretRequiredLowercases,
		RequiredUppercases: secrethelper.DefaultSecretRequiredUppercases,
		RequiredDigits:     secrethelper.DefaultSecretRequiredDigits,
		RequiredSpecials:   secrethelper.DefaultSecretRequiredSpecials,
		MaxSimilarRun:      secrethelper.DefaultSecretMaxSimilarRun,
		MaxSequenceRun:     secrethelper.DefaultSecretMaxSequenceRun,
	}
)

// repoLockPrefix namespaces the advisory lock so it cannot collide with locks taken elsewhere.
//
// NOTE: the value still says "cleanup" because it is the lock key itself. Renaming it would give
// the same repository a different key, so a process still running the old code would not be
// excluded from one running the new code.
const repoLockPrefix = "backup-repo:cleanup:"

// RepoLockName is the advisory lock guarding everything that reconciles a repository against its
// stored records - the cleanup endpoint, the scheduled cleanup job, and the sync endpoint.
//
// Sync has to take it even though it never writes to the repository: it lists the repository and
// then reconciles, so interleaving with a cleanup would let it re-add records for snapshots the
// cleanup had just expired, under fresh IDs.
func RepoLockName(repoSettingID string) string {
	return repoLockPrefix + repoSettingID
}

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

type SyncRepoReq struct {
	Scope       *entity.ObjectScope
	RepoSetting *entity.Setting
	RefObjects  *entity.RefObjects
}

type SyncRepoResp struct {
	// Config is what the repository is configured with right now. These options live inside the
	// repository rather than in the setting, so this is the source of truth whenever they were
	// changed outside the app.
	Config *backup.RepoConfig

	// Snapshots is everything the repository holds, for the caller to reconcile its records
	// against. Unlike a cleanup this is a plain read: nothing in the repository is touched.
	Snapshots []*RepoSnapshot
}

type SyncRepoSnapshotsReq struct {
	Scope       *entity.ObjectScope
	RepoSetting *entity.Setting
	// Remaining is what the repository holds now; stored records are reconciled against it.
	Remaining []*RepoSnapshot
}

type SyncRepoSnapshotsResp struct {
	// Removed carries the snapshots themselves, not just a count, so the caller can report which
	// ones are gone.
	Removed []*entity.BackupSnapshot
	Added   int
}
