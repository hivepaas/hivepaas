package docker

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
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

type VolumeListOption func(*volume.ListOptions)

func (m *manager) VolumeList(
	ctx context.Context,
	options ...VolumeListOption,
) (*volume.ListResponse, error) {
	opts := volume.ListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumeList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) VolumeCreate(
	ctx context.Context,
	options *volume.CreateOptions,
) (*volume.Volume, error) {
	resp, err := m.client.VolumeCreate(ctx, *options)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) VolumeUpdate(
	ctx context.Context,
	volumeID string,
	version *swarm.Version,
	options *volume.UpdateOptions,
) error {
	if options == nil {
		options = &volume.UpdateOptions{}
	}
	if version == nil {
		resp, _, err := m.client.VolumeInspectWithRaw(ctx, volumeID)
		if err != nil {
			return apperrors.NewInfra(err)
		}
		version = &resp.ClusterVolume.Version
	}
	err := m.client.VolumeUpdate(ctx, volumeID, *version, *options)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}

func (m *manager) VolumeRemove(
	ctx context.Context,
	volumeID string,
	force bool,
) error {
	err := m.client.VolumeRemove(ctx, volumeID, force)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}

func (m *manager) VolumeInspect(
	ctx context.Context,
	volumeID string,
) (*volume.Volume, []byte, error) {
	resp, raw, err := m.client.VolumeInspectWithRaw(ctx, volumeID)
	if err != nil {
		return nil, nil, apperrors.NewInfra(err)
	}
	return &resp, raw, nil
}

type VolumesPruneOption func(options *filters.Args)

func (m *manager) VolumesPrune(
	ctx context.Context,
	anonymousOnly bool,
	options ...VolumesPruneOption,
) (*volume.PruneReport, error) {
	opts := filters.Args{}
	if !anonymousOnly {
		FilterAdd(&opts, "all", "true")
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.VolumesPrune(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
