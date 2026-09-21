package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *RegistrySettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRegistrySettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentRegistrySettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so there is nothing to migrate from yet.

	setting.Version = CurrentRegistrySettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
