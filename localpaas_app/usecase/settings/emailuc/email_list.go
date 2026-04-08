package emailuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
)

func (uc *UC) ListEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.ListEmailReq,
) (*emaildto.ListEmailResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := emaildto.TransformEmails(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &emaildto.ListEmailResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
