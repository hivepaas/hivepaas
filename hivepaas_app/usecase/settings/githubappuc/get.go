package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
)

func (uc *UC) GetGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.GetGithubAppReq,
) (*githubappdto.GetGithubAppResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsGithubApp().Decrypt(); err != nil {
			return nil, apperrors.New(err)
		}
	}

	input := &githubappdto.GithubAppTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current.SsoBaseCallbackURL(),
	}
	respData, err := githubappdto.TransformGithubApp(setting, input)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &githubappdto.GetGithubAppResp{
		Data: respData,
	}, nil
}
