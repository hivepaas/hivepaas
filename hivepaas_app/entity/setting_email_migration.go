package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *Email) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentEmailVersion {
		return false, nil
	}
	if setting.Version > CurrentEmailVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentEmailVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
