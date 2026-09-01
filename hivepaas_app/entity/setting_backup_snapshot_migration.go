package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *BackupSnapshot) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentBackupSnapshotVersion {
		return false, nil
	}
	if setting.Version > CurrentBackupSnapshotVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentBackupSnapshotVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
