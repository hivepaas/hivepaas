package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *ClusterVolume) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentClusterVolumeVersion {
		return false, nil
	}
	if setting.Version > CurrentClusterVolumeVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentClusterVolumeVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
