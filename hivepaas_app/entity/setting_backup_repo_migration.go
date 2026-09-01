package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *BackupRepo) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentBackupRepoVersion {
		return false, nil
	}
	if setting.Version > CurrentBackupRepoVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentBackupRepoVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
