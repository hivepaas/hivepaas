package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *ProjectSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentProjectSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentProjectSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentProjectSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
