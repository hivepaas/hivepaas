package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) GetVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.GetVolumeReq,
) (*volumedto.GetVolumeResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listVolumesInDocker(ctx, []*entity.Setting{resp.Data}, nil, refClusterObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := volumedto.TransformVolume(resp.Data, resp.RefObjects, refClusterObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.GetVolumeResp{
		Data: respData,
	}, nil
}
