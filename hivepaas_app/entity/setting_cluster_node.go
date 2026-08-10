package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentClusterNodeVersion = 1
)

var _ = registerSettingParser(base.SettingTypeClusterNode, &clusterNodeParser{})

type clusterNodeParser struct {
}

func (s *clusterNodeParser) New() SettingData {
	return &ClusterNode{}
}

type ClusterNode struct {
	RefID string `json:"refId"`
}

func (s *ClusterNode) GetType() base.SettingType {
	return base.SettingTypeClusterNode
}

func (s *ClusterNode) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ClusterNode) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsClusterNode() (*ClusterNode, error) {
	return parseSettingAs[*ClusterNode](s)
}

func (s *Setting) MustAsClusterNode() *ClusterNode {
	return gofn.Must(s.AsClusterNode())
}
