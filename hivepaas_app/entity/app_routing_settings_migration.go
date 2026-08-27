package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppRoutingSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppRoutingSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppRoutingSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppRoutingSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
