package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *LicenseSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentLicenseSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentLicenseSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	setting.Version = CurrentLicenseSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
