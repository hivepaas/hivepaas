package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *TraefikConfig) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentTraefikConfigVersion {
		return false, nil
	}
	if setting.Version > CurrentTraefikConfigVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentTraefikConfigVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
