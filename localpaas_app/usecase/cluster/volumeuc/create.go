package volumeuc

import (
	"context"
	"fmt"
	"maps"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/dockerhelper"
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
	res, err := uc.dockerManager.VolumeList(ctx, func(options *client.VolumeListOptions) {
		docker.FilterAdd(&options.Filters, "name", req.Name)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if res != nil && len(res.Items) > 0 {
		return nil, apperrors.New(apperrors.ErrInfraAlreadyExists).
			WithNTParam("Error", apperrors.Fmt("Volume '%v' already exists", req.Name))
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

	// Setup labels
	labels := dockerhelper.ApplyUserLabels(map[string]string{}, req.Labels)
	labels[localpaasVolumeLabel] = ""

	if req.ProjectID != "" {
		project, err := uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		labels[docker.StackLabelNamespace] = project.Key
	} else if !req.AvailInProjects {
		labels[docker.StackLabelNamespace] = namespaceGlobal
	}

	createResp, err := uc.dockerManager.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) {
		opts.Driver = string(req.Driver)
		opts.DriverOpts = driverOpts
		opts.Labels = labels
		opts.Name = req.Name
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	volID := createResp.Volume.Name
	if createResp.Volume.ClusterVolume != nil {
		volID = createResp.Volume.ClusterVolume.ID
	}

	return &volumedto.CreateVolumeResp{
		Data: &basedto.ObjectIDResp{ID: volID},
	}, nil
}
