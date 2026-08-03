package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *CloudStorage) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCloudStorageVersion {
		return false, nil
	}
	if setting.Version > CurrentCloudStorageVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCloudStorageVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
