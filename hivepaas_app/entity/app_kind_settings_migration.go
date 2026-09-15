package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"

func (s *AppKindSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppKindSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppKindSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppKindSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
