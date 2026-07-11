package volumeuc

import (
	"context"

	"github.com/moby/moby/api/types/volume"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) ListVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.ListVolumeReq,
) (_ *volumedto.ListVolumeResp, err error) {
	var currVols []volume.Volume
	if req.Scope.IsGlobalScope() {
		currVols, err = uc.clusterService.SyncVolumes(ctx, uc.DB)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listVolumesInDocker(ctx, resp.Data, currVols, refClusterObjects)
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
	currVols []volume.Volume,
	refClusterObjects *entity.RefClusterObjects,
) error {
	if currVols == nil {
		volumes := make([]string, 0, len(settings))
		for _, setting := range settings {
			volumes = append(volumes, dockerhelper.ParseID(setting.ID))
		}
		if len(volumes) == 0 {
			return nil
		}

		res, err := uc.dockerManager.VolumeListByIDs(ctx, volumes)
		if err != nil {
			return apperrors.New(err)
		}
		currVols = res.Items
	}

	for i := range currVols {
		vol := &currVols[i]
		volID := vol.Name
		if vol.ClusterVolume != nil {
			volID = vol.ClusterVolume.ID
		}
		refClusterObjects.RefVolumes[volID] = vol
	}
	return nil
}
