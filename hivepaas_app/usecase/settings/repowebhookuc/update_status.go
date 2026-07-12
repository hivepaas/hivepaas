package repowebhookuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

func (uc *UC) UpdateRepoWebhookStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *repowebhookdto.UpdateRepoWebhookStatusReq,
) (*repowebhookdto.UpdateRepoWebhookStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &repowebhookdto.UpdateRepoWebhookStatusResp{}, nil
}
