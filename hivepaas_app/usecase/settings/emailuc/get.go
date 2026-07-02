package emailuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc/emaildto"
)

func (uc *UC) GetEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.GetEmailReq,
) (*emaildto.GetEmailResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsEmail().Decrypt(); err != nil {
			return nil, apperrors.New(err)
		}
	}

	respData, err := emaildto.TransformEmail(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &emaildto.GetEmailResp{
		Data: respData,
	}, nil
}
