package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *AppFeatureSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppFeatureSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppFeatureSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppFeatureSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
