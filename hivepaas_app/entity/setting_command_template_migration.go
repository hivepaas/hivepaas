package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *CommandTemplate) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCommandTemplateVersion {
		return false, nil
	}
	if setting.Version > CurrentCommandTemplateVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCommandTemplateVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
