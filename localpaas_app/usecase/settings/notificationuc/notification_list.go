package notificationuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) ListNotification(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.ListNotificationReq,
) (*notificationdto.ListNotificationResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := notificationdto.TransformNotifications(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &notificationdto.ListNotificationResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
