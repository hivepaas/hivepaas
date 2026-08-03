package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *DomainSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentDomainSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentDomainSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentDomainSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
