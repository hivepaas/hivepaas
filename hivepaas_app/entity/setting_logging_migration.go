package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *Logging) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentLoggingVersion {
		return false, nil
	}
	if setting.Version > CurrentLoggingVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so there is nothing to migrate from yet.

	setting.Version = CurrentLoggingVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
