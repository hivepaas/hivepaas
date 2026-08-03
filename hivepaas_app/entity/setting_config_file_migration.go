package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *ConfigFile) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentConfigFileVersion {
		return false, nil
	}
	if setting.Version > CurrentConfigFileVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentConfigFileVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
