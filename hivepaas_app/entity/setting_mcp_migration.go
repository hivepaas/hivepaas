package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *MCPSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentMCPSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentMCPSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	setting.Version = CurrentMCPSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
