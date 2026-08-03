package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *ImageBuildSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentImageBuildSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentImageBuildSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentImageBuildSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
