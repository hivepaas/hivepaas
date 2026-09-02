package backupmodel

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// NOTE: these live here rather than in package `backup` because `backup` imports the engine
// packages (kopia, ...), so those packages cannot import it back. `backup` re-exports them.
var (
	// Storage configuration
	ErrStorageConfigRequired = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_BACKUP_STORAGE_CONFIG_REQUIRED")
	ErrStorageTypeRequired   = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_BACKUP_STORAGE_TYPE_REQUIRED")
	ErrEngineUnsupported     = hperrors.NewErr(hperrors.ErrUnsupported, "ERR_BACKUP_ENGINE_UNSUPPORTED")

	// Command execution
	ErrCommandRequired        = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_BACKUP_COMMAND_REQUIRED")
	ErrCommandExecutorMissing = hperrors.NewErr(hperrors.ErrInternal, "ERR_BACKUP_COMMAND_EXECUTOR_MISSING")
	ErrCommandFailed          = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_BACKUP_COMMAND_FAILED")

	// Repository configuration read back from the repository
	ErrRepoConfigUnreadable = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_BACKUP_REPO_CONFIG_UNREADABLE")

	// Snapshots
	ErrSnapshotNotFound        = hperrors.NewErr(hperrors.ErrNotFound, "ERR_BACKUP_SNAPSHOT_NOT_FOUND")
	ErrSnapshotManifestInvalid = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_BACKUP_SNAPSHOT_MANIFEST_INVALID")
)
