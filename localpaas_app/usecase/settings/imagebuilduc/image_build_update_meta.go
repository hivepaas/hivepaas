package imagebuilduc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc/imagebuilddto"
)

func (uc *UC) UpdateImageBuildMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuilddto.UpdateImageBuildMetaReq,
) (*imagebuilddto.UpdateImageBuildMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingMetaData,
			pData *settings.PersistingSettingMetaData,
		) error {
			err := uc.SettingRepo.EnsureUnique(ctx, db, req.Scope, req.Type)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuilddto.UpdateImageBuildMetaResp{}, nil
}
