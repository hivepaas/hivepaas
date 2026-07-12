package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	CurrentClusterVolumeVersion = 1
)

var _ = registerSettingParser(base.SettingTypeClusterVolume, &clusterVolumeParser{})

type clusterVolumeParser struct {
}

func (s *clusterVolumeParser) New() SettingData {
	return &ClusterVolume{}
}

type ClusterVolume struct {
	Driver docker.VolumeDriver `json:"driver"`
}

func (s *ClusterVolume) GetType() base.SettingType {
	return base.SettingTypeClusterVolume
}

func (s *ClusterVolume) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ClusterVolume) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ClusterVolume) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentClusterVolumeVersion {
		return false, nil
	}
	if setting.Version > CurrentClusterVolumeVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentClusterVolumeVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsClusterVolume() (*ClusterVolume, error) {
	return parseSettingAs[*ClusterVolume](s)
}

func (s *Setting) MustAsClusterVolume() *ClusterVolume {
	return gofn.Must(s.AsClusterVolume())
}
