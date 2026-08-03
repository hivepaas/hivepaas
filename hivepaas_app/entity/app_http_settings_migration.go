package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *AppHttpSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppHttpSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppHttpSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppHttpSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
