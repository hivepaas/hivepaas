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

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsNotification().MustDecrypt()
	}

	respData, err := notificationdto.TransformNotification(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &notificationdto.GetNotificationResp{
		Data: respData,
	}, nil
}
