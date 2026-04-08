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

	resp.Data.MustAsEmail().MustDecrypt()
	respData, err := emaildto.TransformEmail(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &emaildto.GetEmailResp{
		Data: respData,
	}, nil
}
