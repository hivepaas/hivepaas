package volumeuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

func (uc *UC) GetVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.GetVolumeReq,
) (*volumedto.GetVolumeResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listVolumesInDocker(ctx, []*entity.Setting{resp.Data}, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := volumedto.TransformVolume(resp.Data, resp.RefObjects, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.GetVolumeResp{
		Data: respData,
	}, nil
}
