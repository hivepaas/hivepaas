package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *KeyAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentKeyAuthVersion {
		return false, nil
	}
	if setting.Version > CurrentKeyAuthVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentKeyAuthVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
