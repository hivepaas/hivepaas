package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *DomainSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentDomainSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentDomainSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentDomainSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
