package imagebuilduc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc/imagebuilddto"
)

func (uc *UC) ListImageBuild(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuilddto.ListImageBuildReq,
) (*imagebuilddto.ListImageBuildResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := imagebuilddto.TransformImageBuilds(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuilddto.ListImageBuildResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
