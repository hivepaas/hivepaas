package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *OAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentOAuthVersion {
		return false, nil
	}
	if setting.Version > CurrentOAuthVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentOAuthVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
