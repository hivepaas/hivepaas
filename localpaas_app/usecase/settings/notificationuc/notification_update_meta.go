package notificationuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) UpdateNotificationMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.UpdateNotificationMetaReq,
) (*notificationdto.UpdateNotificationMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &notificationdto.UpdateNotificationMetaResp{}, nil
}
