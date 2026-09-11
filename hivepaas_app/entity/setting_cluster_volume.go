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
	// Which node the volume's data is on, by id or by label. Decided when the
	// volume is created and never changed afterwards: an immutable answer is
	// what lets a mount spec and a placement constraint derived from it stay
	// correct for as long as the volume exists.
	NodeID    string `json:"nodeId,omitempty"`
	NodeLabel string `json:"nodeLabel,omitempty"`

	// Managed says HivePaaS wrote the specification below and is responsible for
	// reproducing the volume on whatever node needs it. A volume that arrived
	// through discovery belongs to somebody else: it is mounted by name and
	// nothing about it is inferred.
	Managed    bool              `json:"managed,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driverOpts,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
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

// IsPinned reports whether the volume names a node it has to be on.
func (s *ClusterVolume) IsPinned() bool {
	if s == nil {
		return false
	}
	return s.NodeID != "" || s.NodeLabel != ""
}
