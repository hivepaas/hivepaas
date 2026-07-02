package repowebhookuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

func (uc *UC) ListRepoWebhook(
	ctx context.Context,
	auth *basedto.Auth,
	req *repowebhookdto.ListRepoWebhookReq,
) (*repowebhookdto.ListRepoWebhookResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := repowebhookdto.TransformRepoWebhooks(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &repowebhookdto.ListRepoWebhookResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
