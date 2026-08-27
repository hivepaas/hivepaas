package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *PeriodicJob) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentPeriodicJobVersion {
		return false, nil
	}
	if setting.Version > CurrentPeriodicJobVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentPeriodicJobVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
