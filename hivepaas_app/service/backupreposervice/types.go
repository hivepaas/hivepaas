package backupreposervice

import (
	"io"

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

// RepoTarget is the repository a snapshot goes into: its setting, and the scope that
// resolves its storage.
type RepoTarget struct {
	Scope       *entity.ObjectScope
	RepoSetting *entity.Setting
	RefObjects  *entity.RefObjects
}

type BackupStreamReq struct {
	RepoTarget
	Stdin    io.Reader
	FileName string
	// Source is what kopia records the snapshot under, user@host:/path.
	Source      string
	Description string
	Tags        []string
	// Progress, when set, is told the steps the backup takes, such as a
	// repository server started for it.
	Progress func(msg string)
	// OnConnected, when set, is called once the engine is built and connected:
	// from then on the stream uses the database no more, and whatever writes
	// Stdin may use it.
	OnConnected func()
}

type BackupDirectoryReq struct {
	RepoTarget
	// Progress, when set, is told the steps the backup takes.
	Progress func(msg string)
	// HostDir is the directory as the node's host sees it.
	HostDir   string
	NodeID    string
	NodeLabel string
	// Source is what kopia records the snapshot under, user@host:/path.
	Source      string
	Description string
	Tags        []string
}

// OpenRepoServerReq is a repository server to open: the repository, on a
// volume, and the user its clients log in as, user@host.
type OpenRepoServerReq struct {
	RepoTarget
	Username string
}

// RepoServerSession is a repository server running for the caller, until Close.
type RepoServerSession struct {
	URL         string
	Fingerprint string
	// Username and Hostname are the identity clients log in and write snapshots as.
	Username string
	Hostname string
	Password string
	// Stop stops the server; Close calls it once.
	Stop func() error
}

func (s *RepoServerSession) Close() error {
	if s.Stop == nil {
		return nil
	}
	stop := s.Stop
	s.Stop = nil
	return stop()
}

type BackupResp struct {
	Snapshot *RepoSnapshot
}

type DeleteSnapshotReq struct {
	RepoTarget
	SnapshotID string
}

// VolumeHostDir is a volume's directory on its node's host.
type VolumeHostDir struct {
	Dir       string
	NodeID    string
	NodeLabel string
}

type RestoreStreamReq struct {
	RepoTarget
	SnapshotID string
	// FileName is the file of the snapshot a stream backup took.
	FileName string
	Stdout   io.Writer
	// Progress, when set, is told the steps the restore takes.
	Progress func(msg string)
	// OnConnected, when set, is called once the engine is connected: from then on
	// the restore uses the database no more.
	OnConnected func()
}

type RestoreDirectoryReq struct {
	RepoTarget
	SnapshotID string
	// Path is a directory inside the snapshot to restore alone; "" for all of it.
	Path string
	// HostDir is where it goes, on the host of the node, which NodeID or
	// NodeLabel names.
	HostDir   string
	NodeID    string
	NodeLabel string
	// Progress, when set, is told the steps the restore takes.
	Progress func(msg string)
}

type ListEntriesReq struct {
	RepoTarget
	SnapshotID string
	// Path is a directory inside the snapshot; "" for its root.
	Path string
}
