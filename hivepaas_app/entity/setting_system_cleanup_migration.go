package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *SystemCleanup) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSystemCleanupVersion {
		return false, nil
	}
	if setting.Version > CurrentSystemCleanupVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSystemCleanupVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
