package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *AppDeploymentSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppDeploymentSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppDeploymentSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppDeploymentSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
