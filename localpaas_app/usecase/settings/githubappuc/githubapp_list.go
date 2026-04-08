package githubappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
)

func (uc *UC) ListGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.ListGithubAppReq,
) (*githubappdto.ListGithubAppResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &githubappdto.GithubAppTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current.SsoBaseCallbackURL(),
	}
	respData, err := githubappdto.TransformGithubApps(resp.Data, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.ListGithubAppResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
