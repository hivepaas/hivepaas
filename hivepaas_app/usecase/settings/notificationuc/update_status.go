package notificationuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) UpdateNotificationStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.UpdateNotificationStatusReq,
) (*notificationdto.UpdateNotificationStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &notificationdto.UpdateNotificationStatusResp{}, nil
}
