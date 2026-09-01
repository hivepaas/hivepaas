package backup

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
	"github.com/hivepaas/hivepaas/services/backup/kopia"
)

// NewEngine creates a new backup Engine instance based on the specified EngineType and Storage configuration.
func NewEngine(
	engineType EngineType,
	storageCfg *Storage,
	commandExec backupmodel.CommandExecutor,
) (Engine, error) {
	if storageCfg == nil {
		return nil, hperrors.Wrap(backupmodel.ErrStorageConfigRequired)
	}
	if storageCfg.StorageS3 == nil && storageCfg.StorageLocal == nil {
		return nil, hperrors.Wrap(backupmodel.ErrStorageTypeRequired)
	}

	switch engineType {
	case EngineTypeKopia:
		return kopia.NewClient(storageCfg, commandExec), nil
	default:
		return nil, hperrors.Wrap(backupmodel.ErrEngineUnsupported).WithNTParam("Name", engineType)
	}
}
