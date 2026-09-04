package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
			// data.Setting is the stored setting; the request only carries a subset
			// of the fields, so the rest must be carried over before it is replaced.
			current, err := data.Setting.AsGithubApp()
			if err != nil {
				return hperrors.Wrap(err)
			}
			githubApp.CarryOverFrom(current)
			req.KeepMaskedSecrets(githubApp, current)

			err = uc.installGithubAppWebhook(ctx, pData.Setting.ID, githubApp, true)
			if err != nil {
				return hperrors.Wrap(err)
			}
			err = pData.Setting.SetData(githubApp)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &githubappdto.UpdateGithubAppResp{}, nil
}
