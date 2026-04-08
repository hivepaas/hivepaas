package imagebuilduc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc/imagebuilddto"
)

func (uc *UC) DeleteImageBuild(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuilddto.DeleteImageBuildReq,
) (*imagebuilddto.DeleteImageBuildResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuilddto.DeleteImageBuildResp{}, nil
}
