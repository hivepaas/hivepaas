package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) ListVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.ListVolumeReq,
) (*volumedto.ListVolumeResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listVolumesInDocker(ctx, resp.Data, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := volumedto.TransformVolumes(resp.Data, resp.RefObjects, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.ListVolumeResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}

func (uc *UC) listVolumesInDocker(
	ctx context.Context,
	settings []*entity.Setting,
	refClusterObjects *entity.RefClusterObjects,
) error {
	volumes := make([]string, 0, len(settings))
	for _, setting := range settings {
		vol, err := setting.AsClusterVolume()
		if err != nil {
			return apperrors.New(err)
		}
		volumes = append(volumes, vol.VolumeID)
	}
	if len(volumes) == 0 {
		return nil
	}

	if len(volumes) == 1 {
		inspectResp, err := uc.dockerManager.VolumeInspect(ctx, volumes[0])
		if err != nil {
			return apperrors.New(err)
		}
		refClusterObjects.RefVolumes[volumes[0]] = &inspectResp.Volume
		return nil
	}

	res, err := uc.dockerManager.VolumeListByIDs(ctx, volumes)
	if err != nil {
		return apperrors.New(err)
	}

	for i := range res.Items {
		vol := &res.Items[i]
		volID := vol.Name
		if vol.ClusterVolume != nil {
			volID = vol.ClusterVolume.ID
		}
		refClusterObjects.RefVolumes[volID] = vol
	}
	return nil
}
