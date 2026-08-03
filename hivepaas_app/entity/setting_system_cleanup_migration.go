package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *SystemCleanup) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSystemCleanupVersion {
		return false, nil
	}
	if setting.Version > CurrentSystemCleanupVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSystemCleanupVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
