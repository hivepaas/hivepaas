package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *SSLProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLProviderVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLProviderVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLProviderVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
