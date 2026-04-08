package notificationuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) GetNotification(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.GetNotificationReq,
) (*notificationdto.GetNotificationResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsNotification().MustDecrypt()
	respData, err := notificationdto.TransformNotification(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &notificationdto.GetNotificationResp{
		Data: respData,
	}, nil
}
