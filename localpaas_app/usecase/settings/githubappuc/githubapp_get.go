package githubappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
)

func (uc *UC) GetGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.GetGithubAppReq,
) (*githubappdto.GetGithubAppResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsGithubApp().MustDecrypt()
	input := &githubappdto.GithubAppTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current.SsoBaseCallbackURL(),
	}
	respData, err := githubappdto.TransformGithubApp(resp.Data, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.GetGithubAppResp{
		Data: respData,
	}, nil
}
