package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppPlacementSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppPlacementVersion {
		return false, nil
	}
	if setting.Version > CurrentAppPlacementVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppPlacementVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
