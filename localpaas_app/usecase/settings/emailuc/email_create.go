package emailuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
)

const (
	currentSettingType    = base.SettingTypeEmail
	currentSettingVersion = entity.CurrentEmailVersion
)

func (uc *UC) CreateEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.CreateEmailReq,
) (*emaildto.CreateEmailResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:     req.Name,
		DefaultMustUnique: true,
		Version:           currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			err := pData.Setting.SetData(req.ToEntity())
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &emaildto.CreateEmailResp{
		Data: resp.Data,
	}, nil
}
