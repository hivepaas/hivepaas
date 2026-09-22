package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppDeploymentSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppDeploymentSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppDeploymentSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 2 removed repoSource.imageName, repoSource.imageTags and noCache.
	// The name is a function of the app now, the tags were configuration no build
	// ever read, and noCache belongs to one deployment rather than to the app.
	// All three are gone from the struct, so parsing has already dropped them -
	// writing the data back is what takes them out of the row.

	setting.Version = CurrentAppDeploymentSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
