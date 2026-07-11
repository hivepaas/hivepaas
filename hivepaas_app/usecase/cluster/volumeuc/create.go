package volumeuc

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) CreateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.CreateVolumeReq,
) (*volumedto.CreateVolumeResp, error) {
	req.Type = currentSettingType
	volEntity := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingRefIDs: volEntity.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if req.Scope.IsProjectScope() {
				req.Name = data.ScopeProject.Key + "_" + req.Name
			}
			createResp, err := uc.createVolumeInDocker(ctx, req.VolumeBaseReq)
			if err != nil {
				return apperrors.New(err)
			}
			vol := &createResp.Volume
			volID := vol.Name
			if vol.ClusterVolume != nil {
				volID = vol.ClusterVolume.ID
			}

			pData.Setting.ID = dockerhelper.WrapVolumeID(volID)
			pData.Setting.Name = req.Name
			pData.Setting.Kind = vol.Driver
			if err := pData.Setting.SetData(volEntity); err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.CreateVolumeResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) createVolumeInDocker(
	ctx context.Context,
	req *volumedto.VolumeBaseReq,
) (*client.VolumeCreateResult, error) {
	_, err := uc.dockerManager.VolumeInspect(ctx, req.Name)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.New(err)
	}
	if err == nil {
		return nil, apperrors.NewAlreadyExist("Cluster volume")
	}

	driverOpts := map[string]string{}
	if req.Driver == docker.VolumeDriverLocal { //nolint:nestif
		switch {
		case req.BindOptions != nil:
			driverOpts["type"] = "none"
			driverOpts["device"] = req.BindOptions.Directory
			o := fmt.Sprintf("bind,%s", gofn.If(req.BindOptions.Readonly, "ro", "rw"))
			if req.BindOptions.Propagation != "" {
				o += "," + string(req.BindOptions.Propagation)
			}
			if req.BindOptions.ExtraOptions != "" {
				o += "," + req.BindOptions.ExtraOptions
			}
			driverOpts["o"] = o

		case req.NfsOptions != nil:
			driverOpts["type"] = "nfs"
			driverOpts["device"] = req.NfsOptions.Device
			o := fmt.Sprintf("addr=%s,%s", req.NfsOptions.Addr, gofn.If(req.NfsOptions.Readonly, "ro", "rw"))
			if req.NfsOptions.Version != "" {
				o += ",nfsvers=" + req.NfsOptions.Version
			}
			if req.NfsOptions.ExtraOptions != "" {
				o += "," + req.NfsOptions.ExtraOptions
			}
			driverOpts["o"] = o

		case req.TmpfsOptions != nil:
			driverOpts["type"] = "tmpfs"
			driverOpts["device"] = gofn.Coalesce(req.TmpfsOptions.Device, "tmpfs")
			bytes := req.TmpfsOptions.Size.Bytes() + int64(unit.MB) - 1
			o := fmt.Sprintf("size=%vm", bytes/int64(unit.MB))
			if req.TmpfsOptions.Mode > 0 {
				o += fmt.Sprintf(",mode=%v", req.TmpfsOptions.Mode)
			}
			if req.TmpfsOptions.UID > 0 {
				o += fmt.Sprintf(",uid=%v", req.TmpfsOptions.UID)
			}
			if req.TmpfsOptions.GID > 0 {
				o += fmt.Sprintf(",gid=%v", req.TmpfsOptions.GID)
			}
			driverOpts["o"] = o

		case req.BtrfsOptions != nil:
			driverOpts["type"] = "btrfs"
			driverOpts["device"] = req.BtrfsOptions.Device
		}
	}
	// Overwrite the driver opts with the extra values from the client
	maps.Copy(driverOpts, req.Options)

	createResp, err := uc.dockerManager.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) {
		opts.Driver = string(req.Driver)
		opts.DriverOpts = driverOpts
		opts.Labels = req.Labels
		opts.Name = req.Name
	})
	if err != nil {
		return nil, apperrors.New(err)
	}
	return createResp, nil
}
