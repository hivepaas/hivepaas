package docker

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type VolumeDriver string

const (
	VolumeDriverLocal VolumeDriver = "local"
)

// VolumeNameLabel carries the name a person gave the volume, because its docker
// name is a ULID.
const VolumeNameLabel = "hivepaas.volume.name"

type VolumeScope string

const (
	VolumeScopeGlobal VolumeScope = "global"
	VolumeScopeLocal  VolumeScope = "local"
)

var (
	AllVolumeScopes = []VolumeScope{VolumeScopeGlobal, VolumeScopeLocal}
)

type VolumeType string

const (
	VolumeTypeVolume  VolumeType = "volume"
	VolumeTypeCluster VolumeType = "cluster"
)

var (
	AllVolumeTypes = []VolumeType{VolumeTypeVolume, VolumeTypeCluster}
)

type VolumeListOption func(*client.VolumeListOptions)

func (m *manager) VolumeList(
	ctx context.Context,
	options ...VolumeListOption,
) (*client.VolumeListResult, error) {
	opts := client.VolumeListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumeList(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

// VolumeListByIDs lists the volumes of these ids, in one call, and only them,
// an id being what VolumeID says. Docker has no filter for a cluster volume's
// id, so every volume is listed and narrowed here; see byids.go.
func (m *manager) VolumeListByIDs(
	ctx context.Context,
	volumeIDs []string,
	options ...VolumeListOption,
) (*client.VolumeListResult, error) {
	wanted := wantedIDs(volumeIDs)
	if len(wanted) == 0 {
		return &client.VolumeListResult{Items: []volume.Volume{}}, nil
	}
	resp, err := m.VolumeList(ctx, options...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Items = keepWanted(resp.Items, wanted, VolumeID)
	return resp, nil
}

// VolumeID is a volume's id: its cluster id for a cluster volume, its name for
// any other.
func VolumeID(vol *volume.Volume) string {
	if vol == nil {
		return ""
	}
	if vol.ClusterVolume == nil {
		return vol.Name
	}
	return vol.ClusterVolume.ID
}

type VolumeCreateOption func(options *client.VolumeCreateOptions)

func (m *manager) VolumeCreate(
	ctx context.Context,
	options ...VolumeCreateOption,
) (*client.VolumeCreateResult, error) {
	opts := client.VolumeCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumeCreate(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) VolumeUpdate(
	ctx context.Context,
	volumeID string,
	version *swarm.Version,
	spec *volume.ClusterVolumeSpec,
) (*client.VolumeUpdateResult, error) {
	if spec == nil {
		return nil, nil
	}

	opts := client.VolumeUpdateOptions{
		Spec: spec,
	}

	if version == nil {
		resp, err := m.VolumeInspect(ctx, volumeID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		version = &resp.Volume.ClusterVolume.Version
	}
	opts.Version = *version

	resp, err := m.client.VolumeUpdate(ctx, volumeID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type VolumeRemoveOption func(options *client.VolumeRemoveOptions)

func (m *manager) VolumeRemove(
	ctx context.Context,
	volumeID string,
	force bool,
	options ...VolumeRemoveOption,
) (*client.VolumeRemoveResult, error) {
	opts := client.VolumeRemoveOptions{}
	opts.Force = force
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumeRemove(ctx, volumeID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type VolumeInspectOption func(options *client.VolumeInspectOptions)

func (m *manager) VolumeInspect(
	ctx context.Context,
	volumeID string,
	options ...VolumeInspectOption,
) (*client.VolumeInspectResult, error) {
	opts := client.VolumeInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumeInspect(ctx, volumeID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type VolumePruneOption func(options *client.VolumePruneOptions)

func (m *manager) VolumePrune(
	ctx context.Context,
	anonymousOnly bool,
	options ...VolumePruneOption,
) (*client.VolumePruneResult, error) {
	opts := client.VolumePruneOptions{}
	if !anonymousOnly {
		FilterAdd(&opts.Filters, "all", "true")
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumePrune(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}
