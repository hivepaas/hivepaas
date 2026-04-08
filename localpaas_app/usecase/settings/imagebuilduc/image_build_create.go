package imagebuilduc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc/imagebuilddto"
)

const (
	currentSettingType    = base.SettingTypeImageBuild
	currentSettingVersion = entity.CurrentImageBuildVersion

	defaultName = "image build settings"
)

func (uc *UC) CreateImageBuild(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuilddto.CreateImageBuildReq,
) (*imagebuilddto.CreateImageBuildResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		Version: currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Name = defaultName
			err := pData.Setting.SetData(req.ToEntity())
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
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

	return &imagebuilddto.CreateImageBuildResp{
		Data: resp.Data,
	}, nil
}
