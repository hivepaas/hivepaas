package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *IMService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentIMServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentIMServiceVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentIMServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
