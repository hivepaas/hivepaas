package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *EnvVars) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentEnvVarsVersion {
		return false, nil
	}
	if setting.Version > CurrentEnvVarsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentEnvVarsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
