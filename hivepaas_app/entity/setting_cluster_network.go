package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentClusterNetworkVersion = 1
)

var _ = registerSettingParser(base.SettingTypeClusterNetwork, &clusterNetworkParser{})

type clusterNetworkParser struct {
}

func (s *clusterNetworkParser) New() SettingData {
	return &ClusterNetwork{}
}

type ClusterNetwork struct {
}

func (s *ClusterNetwork) GetType() base.SettingType {
	return base.SettingTypeClusterNetwork
}

func (s *ClusterNetwork) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ClusterNetwork) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ClusterNetwork) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentClusterNetworkVersion {
		return false, nil
	}
	if setting.Version > CurrentClusterNetworkVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentClusterNetworkVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsClusterNetwork() (*ClusterNetwork, error) {
	return parseSettingAs[*ClusterNetwork](s)
}

func (s *Setting) MustAsClusterNetwork() *ClusterNetwork {
	return gofn.Must(s.AsClusterNetwork())
}
