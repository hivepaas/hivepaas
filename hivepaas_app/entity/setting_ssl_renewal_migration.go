package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *SSLRenewal) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLRenewalVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLRenewalVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLRenewalVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
