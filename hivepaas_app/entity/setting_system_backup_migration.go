package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *SystemBackup) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSystemBackupVersion {
		return false, nil
	}
	if setting.Version > CurrentSystemBackupVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSystemBackupVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
