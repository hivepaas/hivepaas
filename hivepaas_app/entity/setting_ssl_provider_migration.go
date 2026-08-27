package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *SSLProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLProviderVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLProviderVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLProviderVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
