package repowebhookuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

func (uc *UC) CreateRepoWebhook(
	ctx context.Context,
	auth *basedto.Auth,
	req *repowebhookdto.CreateRepoWebhookReq,
) (*repowebhookdto.CreateRepoWebhookResp, error) {
	req.Type = currentSettingType
	webhookData := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: webhookData.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			if webhookData.Secret == "" { // generate secret if empty
				webhookData.Secret = gofn.RandTokenAsHex(base.DefaultWebhookSecretByteLen)
			}
			err := pData.Setting.SetData(webhookData)
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &repowebhookdto.CreateRepoWebhookResp{
		Data: &repowebhookdto.RepoWebhookDataResp{
			ID:         resp.Data.ID,
			Secret:     webhookData.Secret,
			WebhookURL: config.Current.RepoWebhookURL(resp.Data.ID),
		},
	}, nil
}
