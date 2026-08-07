package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *Script) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentScriptVersion {
		return false, nil
	}
	if setting.Version > CurrentScriptVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentScriptVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
