package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"

func (s *AppTemplateSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppTemplateSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppTemplateSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	setting.Version = CurrentAppTemplateSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
