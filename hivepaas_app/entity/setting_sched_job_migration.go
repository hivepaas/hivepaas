package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *SchedJob) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSchedJobVersion {
		return false, nil
	}
	if setting.Version > CurrentSchedJobVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSchedJobVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
