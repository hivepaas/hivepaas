package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppCloneSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppCloneSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppCloneSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppCloneSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
