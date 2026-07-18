package entity

import (
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func NewRefClusterObjects() *RefClusterObjects {
	return &RefClusterObjects{
		RefNodes:    make(map[string]*swarm.Node),
		RefVolumes:  make(map[string]*volume.Volume),
		RefNetworks: make(map[string]*network.Network),
	}
}

type RefClusterObjects struct {
	RefNodes    map[string]*swarm.Node
	RefVolumes  map[string]*volume.Volume
	RefNetworks map[string]*network.Network
}

func (r *RefClusterObjects) AddRefClusterObjects(refObjects *RefClusterObjects) {
	if refObjects == nil {
		return
	}

	if r.RefNodes == nil {
		r.RefNodes = make(map[string]*swarm.Node, len(refObjects.RefNodes))
	}
	for _, refNode := range refObjects.RefNodes {
		r.RefNodes[refNode.ID] = refNode
	}

	if r.RefVolumes == nil {
		r.RefVolumes = make(map[string]*volume.Volume, len(refObjects.RefVolumes))
	}
	for _, refVolume := range refObjects.RefVolumes {
		volID := refVolume.Name
		if refVolume.ClusterVolume != nil {
			volID = refVolume.ClusterVolume.ID
		}
		r.RefVolumes[volID] = refVolume
	}

	if r.RefNetworks == nil {
		r.RefNetworks = make(map[string]*network.Network, len(refObjects.RefNetworks))
	}
	for _, refUser := range refObjects.RefNetworks {
		r.RefNetworks[refUser.ID] = refUser
	}
}

type RefClusterObjectIDs struct {
	RefNodeIDs    []string
	RefVolumeIDs  []string
	RefNetworkIDs []string
}

func (r *RefClusterObjectIDs) HasData() bool {
	return len(r.RefNodeIDs) > 0 || len(r.RefVolumeIDs) > 0 || len(r.RefNetworkIDs) > 0
}

func (r *RefClusterObjectIDs) AddRefIDs(refIDs *RefClusterObjectIDs) {
	if refIDs == nil {
		return
	}
	r.RefNodeIDs = append(r.RefNodeIDs, refIDs.RefNodeIDs...)
	r.RefVolumeIDs = append(r.RefVolumeIDs, refIDs.RefVolumeIDs...)
	r.RefNetworkIDs = append(r.RefNetworkIDs, refIDs.RefNetworkIDs...)
}

func (r *RefClusterObjectIDs) GetResourceLinks(srcType base.ResourceType, srcID string) []*ResLink {
	resLinks := make([]*ResLink, 0, len(r.RefNodeIDs)+len(r.RefVolumeIDs)+len(r.RefNetworkIDs))
	timeNow := timeutil.NowUTC()
	for _, refNodeID := range r.RefNodeIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeClusterNode,
			DstID:     refNodeID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refVolumeID := range r.RefVolumeIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeClusterVolume,
			DstID:     refVolumeID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refNetID := range r.RefNetworkIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeClusterNetwork,
			DstID:     refNetID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	return resLinks
}
