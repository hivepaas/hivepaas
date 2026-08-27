package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *Secret) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSecretVersion {
		return false, nil
	}
	if setting.Version > CurrentSecretVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSecretVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
