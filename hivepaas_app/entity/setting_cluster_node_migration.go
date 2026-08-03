package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *ClusterNode) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentClusterNodeVersion {
		return false, nil
	}
	if setting.Version > CurrentClusterNodeVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentClusterNodeVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
