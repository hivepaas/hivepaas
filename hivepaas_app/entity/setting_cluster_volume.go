package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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
	RefID string `json:"refId"`
}

func (s *ClusterVolume) GetType() base.SettingType {
	return base.SettingTypeClusterVolume
}

func (s *ClusterVolume) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ClusterVolume) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsClusterVolume() (*ClusterVolume, error) {
	return parseSettingAs[*ClusterVolume](s)
}

func (s *Setting) MustAsClusterVolume() *ClusterVolume {
	return gofn.Must(s.AsClusterVolume())
}
