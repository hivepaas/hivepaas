package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *SSLCert) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLCertVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLCertVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLCertVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
