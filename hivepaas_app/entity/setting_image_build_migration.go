package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *ImageBuildSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentImageBuildVersion {
		return false, nil
	}
	if setting.Version > CurrentImageBuildVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentImageBuildVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
