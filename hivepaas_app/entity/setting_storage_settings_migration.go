package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *StorageSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentStorageSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentStorageSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentStorageSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
