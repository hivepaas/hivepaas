package volumeuc

import (
	"context"
	"fmt"
	"maps"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	namespaceGlobal = "global"
)

func (uc *UC) CreateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.CreateVolumeReq,
) (*volumedto.CreateVolumeResp, error) {
	res, err := uc.dockerManager.VolumeList(ctx, func(options *volume.ListOptions) {
		options.Filters = filters.NewArgs(filters.Arg("name", req.Name))
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(res.Volumes) > 0 {
		return nil, apperrors.New(apperrors.ErrInfraAlreadyExists).
			WithNTParam("Error", fmt.Sprintf("volume '%s' already exists", req.Name))
	}

	driverOpts := map[string]string{}
	switch req.Driver {
	case base.VolumeDriverLocal:
		switch req.Type {
		case base.VolumeTypeVolume:
			// Do nothing

		case base.VolumeTypeNfs:
			driverOpts["type"] = string(req.Type)
			o := fmt.Sprintf("addr=%s,%s", req.NfsOpts.Addr, gofn.If(req.NfsOpts.Readonly, "ro", "rw"))
			if req.NfsOpts.Version != "" {
				o += "," + req.NfsOpts.Version
			}
			driverOpts["o"] = o
			driverOpts["device"] = req.NfsOpts.Device

		default:
			return nil, apperrors.New(apperrors.ErrUnsupported).
				WithMsgLog("driver '%s' does not support volume type '%s'", req.Driver, req.Type)
		}

	case base.VolumeDriverSeaweedFs:
		// TODO: add supports

	default:
		return nil, apperrors.New(apperrors.ErrUnsupported).
			WithMsgLog("driver '%s' is not supported", req.Driver)
	}
	// Overwrite the driver opts with the extra values from the client
	maps.Copy(driverOpts, req.ExtraDriverOpts)

	// Setup default labels
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}
	req.Labels[localpaasVolumeLabel] = ""

	if req.ProjectID != "" {
		project, err := uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		req.Labels[docker.StackLabelNamespace] = project.Key
	} else if !req.AvailInProjects {
		req.Labels[docker.StackLabelNamespace] = namespaceGlobal
	}

	options := &volume.CreateOptions{
		Driver:     string(req.Driver),
		DriverOpts: driverOpts,
		Labels:     req.Labels,
		Name:       req.Name,
	}
	vol, err := uc.dockerManager.VolumeCreate(ctx, options)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	volID := vol.Name
	if vol.ClusterVolume != nil {
		volID = vol.ClusterVolume.ID
	}

	return &volumedto.CreateVolumeResp{
		Data: &basedto.ObjectIDResp{ID: volID},
	}, nil
}
