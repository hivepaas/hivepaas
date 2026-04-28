package volumeuc

import (
	"context"
	"fmt"
	"maps"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
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
	case docker.VolumeDriverLocal:
		switch {
		case req.NfsOptions != nil:
			driverOpts["type"] = "nfs"
			o := fmt.Sprintf("addr=%s,%s", req.NfsOptions.Addr, gofn.If(req.NfsOptions.Readonly, "ro", "rw"))
			if req.NfsOptions.Version != "" {
				o += "," + req.NfsOptions.Version
			}
			driverOpts["o"] = o
			driverOpts["device"] = req.NfsOptions.Device

		case req.TmpfsOptions != nil:
			driverOpts["type"] = "tmpfs"
			bytes := req.TmpfsOptions.Size.Bytes() + int64(unit.MB) - 1
			driverOpts["o"] = fmt.Sprintf("size=%vm,uid=%v", bytes/int64(unit.MB), req.TmpfsOptions.UID)
			driverOpts["device"] = gofn.Coalesce(req.TmpfsOptions.Device, "tmpfs")

		case req.BtrfsOptions != nil:
			driverOpts["type"] = "btrfs"
			driverOpts["device"] = req.BtrfsOptions.Device
		}

	default:
		return nil, apperrors.New(apperrors.ErrUnsupported).
			WithMsgLog("driver '%s' is not supported", req.Driver)
	}
	// Overwrite the driver opts with the extra values from the client
	maps.Copy(driverOpts, req.Options)

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
