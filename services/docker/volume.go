package docker

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type VolumeDriver string

const (
	VolumeDriverLocal VolumeDriver = "local"
)

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
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) VolumeListByIDs(
	ctx context.Context,
	volumeIDOrNames []string,
	options ...VolumeListOption,
) (*client.VolumeListResult, error) {
	resp := &client.VolumeListResult{}
	if len(volumeIDOrNames) == 0 {
		return resp, nil
	}

	if len(volumeIDOrNames) == 1 {
		inspect, err := m.VolumeInspect(ctx, volumeIDOrNames[0])
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.Wrap(err)
		}
		if inspect != nil {
			resp.Items = []volume.Volume{inspect.Volume}
		}
		return resp, nil
	}

	listResp, err := m.VolumeList(ctx, options...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	for i := range listResp.Items {
		vol := &listResp.Items[i]
		if gofn.Contain(volumeIDOrNames, vol.Name) {
			resp.Items = append(resp.Items, *vol)
			continue
		}
		if vol.ClusterVolume != nil && gofn.Contain(volumeIDOrNames, vol.ClusterVolume.ID) {
			resp.Items = append(resp.Items, *vol)
			continue
		}
	}

	return resp, nil
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
		return nil, apperrors.NewInfra(err)
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
			return nil, apperrors.Wrap(err)
		}
		version = &resp.Volume.ClusterVolume.Version
	}
	opts.Version = *version

	resp, err := m.client.VolumeUpdate(ctx, volumeID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
