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

	respData, err := repowebhookdto.TransformRepoWebhook(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &repowebhookdto.GetRepoWebhookResp{
		Data: respData,
	}, nil
}
