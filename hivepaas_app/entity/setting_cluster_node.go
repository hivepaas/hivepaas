package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
}

func (s *ClusterNode) GetType() base.SettingType {
	return base.SettingTypeClusterNode
}

func (s *ClusterNode) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ClusterNode) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ClusterNode) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentClusterNodeVersion {
		return false, nil
	}
	if setting.Version > CurrentClusterNodeVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentClusterNodeVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsClusterNode() (*ClusterNode, error) {
	return parseSettingAs[*ClusterNode](s)
}

func (s *Setting) MustAsClusterNode() *ClusterNode {
	return gofn.Must(s.AsClusterNode())
}
