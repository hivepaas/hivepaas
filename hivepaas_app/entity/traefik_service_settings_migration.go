package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *TraefikService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentTraefikServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentTraefikServiceVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentTraefikServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
