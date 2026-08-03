package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *AcmeDnsProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAcmeDnsProviderVersion {
		return false, nil
	}
	if setting.Version > CurrentAcmeDnsProviderVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAcmeDnsProviderVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
