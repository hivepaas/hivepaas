package backup

import (
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// Re-export errors defined by the engine model, so callers of this package do not have to reach
// into backupmodel for them.
var (
	ErrStorageConfigRequired = backupmodel.ErrStorageConfigRequired
	ErrStorageTypeRequired   = backupmodel.ErrStorageTypeRequired
	ErrEngineUnsupported     = backupmodel.ErrEngineUnsupported

	ErrCommandRequired        = backupmodel.ErrCommandRequired
	ErrCommandExecutorMissing = backupmodel.ErrCommandExecutorMissing
	ErrCommandFailed          = backupmodel.ErrCommandFailed

	ErrRepoConfigUnreadable = backupmodel.ErrRepoConfigUnreadable

	ErrSnapshotNotFound        = backupmodel.ErrSnapshotNotFound
	ErrSnapshotManifestInvalid = backupmodel.ErrSnapshotManifestInvalid
)
