package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *APIKey) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAPIKeyVersion {
		return false, nil
	}
	if setting.Version > CurrentAPIKeyVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAPIKeyVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
