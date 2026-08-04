package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *AppCloneSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppCloneSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppCloneSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppCloneSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
