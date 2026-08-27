package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *ClusterNetwork) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentClusterNetworkVersion {
		return false, nil
	}
	if setting.Version > CurrentClusterNetworkVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentClusterNetworkVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
