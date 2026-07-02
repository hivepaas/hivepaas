package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
)

func (uc *UC) UpdateGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.UpdateGithubAppReq,
) (*githubappdto.UpdateGithubAppResp, error) {
	req.Type = currentSettingType
	githubApp := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: githubApp.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := uc.installGithubAppWebhook(ctx, pData.Setting.ID, githubApp, true)
			if err != nil {
				return apperrors.New(err)
			}
			err = pData.Setting.SetData(githubApp)
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &githubappdto.UpdateGithubAppResp{}, nil
}
