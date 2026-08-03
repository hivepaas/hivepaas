package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *HivePaaSService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentHivePaaSServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentHivePaaSServiceVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentHivePaaSServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
