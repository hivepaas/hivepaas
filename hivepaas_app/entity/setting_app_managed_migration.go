package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppManagedSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppManagedVersion {
		return false, nil
	}
	if setting.Version > CurrentAppManagedVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so an older row is one written before versions
	// were set: the data is already in its shape.
	setting.Version = CurrentAppManagedVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
