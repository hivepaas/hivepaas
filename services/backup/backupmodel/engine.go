package backupmodel

import (
	"context"
	"io"
)

// Engine defines the unified contract that all backup drivers must implement.
type Engine interface {
	// Name returns the identifier of the engine (e.g. "restic", "kopia").
	Name() EngineType

	// InitRepo initializes a new backup repository on the storage backend.
	InitRepo(ctx context.Context, opts *InitRepoOptions) error

	// ApplyRepoOptions applies the changeable settings to an existing repository.
	ApplyRepoOptions(ctx context.Context, opts *RepoOptions) error

	// ReadRepoConfig reads back the settings the repository is actually running with.
	ReadRepoConfig(ctx context.Context) (RepoConfig, error)

	// RefreshCache drops what this client cached about the repository, so the next read goes to
	// the repository itself. Needed before reading back a repository that something else may have
	// reconfigured: an already-connected client otherwise keeps reporting what it cached.
	RefreshCache(ctx context.Context) error

	// ConnectRepo connects to an existing backup repository on the storage backend.
	ConnectRepo(ctx context.Context) error

	// CheckRepo verifies the integrity of the backup repository.
	CheckRepo(ctx context.Context) error

	// ChangePassword changes the repository encryption password.
	ChangePassword(ctx context.Context, oldPassword, newPassword string) error

	// BackupDirectory creates a snapshot from a directory on the local disk.
	BackupDirectory(ctx context.Context, dirPath string, opts *BackupOptions) (
		BackupResult, error)

	// BackupStream creates a snapshot by streaming data directly from an io.Reader (e.g. pg_dump / mysqldump).
	BackupStream(ctx context.Context, stdin io.Reader, filename string, opts *BackupOptions) (
		BackupResult, error)

	// RestoreDirectory restores a snapshot's contents to a target local directory.
	RestoreDirectory(ctx context.Context, snapshotID string, targetDir string, opts *RestoreOptions) (
		RestoreResult, error)

	// RestoreStream restores a single file from a snapshot and streams it to an io.Writer (e.g. pipe to psql).
	RestoreStream(ctx context.Context, snapshotID string, filename string, stdout io.Writer, opts *RestoreOptions) (
		RestoreResult, error)

	// ListSnapshots retrieves snapshots matching the specified options. If opts is nil, returns all snapshots.
	ListSnapshots(ctx context.Context, opts *ListSnapshotsOptions) (
		ListSnapshotsResult, error)

	// GetSnapshot retrieves details of a specific snapshot by its ID.
	GetSnapshot(ctx context.Context, snapshotID string) (
		GetSnapshotResult, error)

	// DeleteSnapshot removes a specific snapshot from the repository.
	DeleteSnapshot(ctx context.Context, snapshotID string) (
		DeleteSnapshotResult, error)

	// Prune applies the retention policy and purges old unreferenced data blocks.
	Prune(ctx context.Context, policy *RetentionPolicy) (
		PruneResult, error)
}
