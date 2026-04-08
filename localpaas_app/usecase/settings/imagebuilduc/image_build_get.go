package imagebuilduc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc/imagebuilddto"
)

func (uc *UC) GetImageBuild(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuilddto.GetImageBuildReq,
) (*imagebuilddto.GetImageBuildResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := imagebuilddto.TransformImageBuild(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuilddto.GetImageBuildResp{
		Data: respData,
	}, nil
}
