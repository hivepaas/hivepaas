package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *RegistryAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRegistryAuthVersion {
		return false, nil
	}
	if setting.Version > CurrentRegistryAuthVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentRegistryAuthVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
