package emailuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
)

func (uc *UC) UpdateEmailMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.UpdateEmailMetaReq,
) (*emaildto.UpdateEmailMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{
		DefaultMustUnique: true,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &emaildto.UpdateEmailMetaResp{}, nil
}
