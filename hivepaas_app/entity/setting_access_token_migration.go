package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *AccessToken) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAccessTokenVersion {
		return false, nil
	}
	if setting.Version > CurrentAccessTokenVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAccessTokenVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
