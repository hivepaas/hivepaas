package emailuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
)

func (uc *UC) GetEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.GetEmailReq,
) (*emaildto.GetEmailResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsEmail().MustDecrypt()
	}

	respData, err := emaildto.TransformEmail(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &emaildto.GetEmailResp{
		Data: respData,
	}, nil
}
