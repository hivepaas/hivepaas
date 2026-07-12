package repowebhookuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

func (uc *UC) DeleteRepoWebhook(
	ctx context.Context,
	auth *basedto.Auth,
	req *repowebhookdto.DeleteRepoWebhookReq,
) (*repowebhookdto.DeleteRepoWebhookResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &repowebhookdto.DeleteRepoWebhookResp{}, nil
}
