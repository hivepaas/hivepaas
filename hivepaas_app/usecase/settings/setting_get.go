package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type GetSettingReq struct {
	BaseSettingReq
	ID string `json:"-" mapstructure:"-"`
}

func (req *GetSettingReq) Validate() (validators []vld.Validator) {
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return
}

type GetSettingResp struct {
	Data       *entity.Setting
	RefObjects *entity.RefObjects
}

type GetSettingData struct {
	BaseSettingData

	SkipLoadingRefObjects bool
	ExtraLoadOpts         []bunex.SelectQueryOption

	AfterLoading func(context.Context, database.IDB, *GetSettingData) error
}

func (uc *BaseUC) GetSetting(
	ctx context.Context,
	auth *basedto.Auth,
	req *GetSettingReq,
	data *GetSettingData,
) (*GetSettingResp, error) {
	db := uc.DB

	err := uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if data.AfterLoading != nil {
		if err = data.AfterLoading(ctx, db, data); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	setting, err := uc.loadSettingByID(ctx, db, &req.BaseSettingReq, req.ID,
		false, data.ExtraLoadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if setting != nil {
		setting.CurrentObjectID = req.Scope.MainObjectID()
	}

	var refObjects *entity.RefObjects
	if !data.SkipLoadingRefObjects {
		refObjects, err = uc.SettingService.LoadReferenceObjects(ctx, db, req.Scope,
			true, false, setting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &GetSettingResp{
		Data:       setting,
		RefObjects: refObjects,
	}, nil
}

func (uc *BaseUC) GetSettingByID(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	id string,
	requireActive bool,
	extraLoadOpts ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	setting, err := uc.loadSettingByID(ctx, db, req, id, requireActive, extraLoadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if setting != nil {
		setting.CurrentObjectID = req.Scope.MainObjectID()
	}
	return setting, nil
}
