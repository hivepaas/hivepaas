package repowebhookuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

func (uc *UC) GetRepoWebhook(
	ctx context.Context,
	auth *basedto.Auth,
	req *repowebhookdto.GetRepoWebhookReq,
) (*repowebhookdto.GetRepoWebhookResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsRepoWebhook().MustDecrypt()
	}

	respData, err := repowebhookdto.TransformRepoWebhook(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &repowebhookdto.GetRepoWebhookResp{
		Data: respData,
	}, nil
}
