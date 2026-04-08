package githubappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
)

func (uc *UC) UpdateGithubAppMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.UpdateGithubAppMetaReq,
) (*githubappdto.UpdateGithubAppMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.UpdateGithubAppMetaResp{}, nil
}
