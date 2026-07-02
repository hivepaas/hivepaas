package notificationuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) DeleteNotification(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.DeleteNotificationReq,
) (*notificationdto.DeleteNotificationResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &notificationdto.DeleteNotificationResp{}, nil
}
