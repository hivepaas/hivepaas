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

	// Version 2 added SystemAppsSync, on for every installation.
	if setting.Version < 2 && s.SystemAppsSync == nil { //nolint:mnd
		s.SystemAppsSync = &SystemAppsSync{Enabled: true}
	}

	setting.Version = CurrentSystemCleanupVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
