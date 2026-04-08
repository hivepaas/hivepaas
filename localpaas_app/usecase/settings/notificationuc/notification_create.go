package notificationuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
)

const (
	currentSettingType    = base.SettingTypeNotification
	currentSettingVersion = entity.CurrentNotificationVersion
)

func (uc *UC) CreateNotification(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.CreateNotificationReq,
) (*notificationdto.CreateNotificationResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName: req.Name,
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			err := pData.Setting.SetData(req.ToEntity())
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &notificationdto.CreateNotificationResp{
		Data: resp.Data,
	}, nil
}
