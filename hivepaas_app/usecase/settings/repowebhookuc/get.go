package repowebhookuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

func (uc *UC) GetRepoWebhook(
	ctx context.Context,
	auth *basedto.Auth,
	req *repowebhookdto.GetRepoWebhookReq,
) (*repowebhookdto.GetRepoWebhookResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsRepoWebhook().Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	respData, err := repowebhookdto.TransformRepoWebhook(setting, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &repowebhookdto.GetRepoWebhookResp{
		Data: respData,
	}, nil
}
