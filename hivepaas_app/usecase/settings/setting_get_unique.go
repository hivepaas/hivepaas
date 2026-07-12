package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type GetUniqueSettingReq struct {
	BaseSettingReq
}

func (req *GetUniqueSettingReq) Validate() (validators []vld.Validator) {
	return
}

type GetUniqueSettingResp struct {
	Data       *entity.Setting
	RefObjects *entity.RefObjects
}

type GetUniqueSettingData struct {
	BaseSettingData

	ExtraLoadOpts []bunex.SelectQueryOption
}

func (uc *BaseUC) GetUniqueSetting(
	ctx context.Context,
	auth *basedto.Auth,
	req *GetUniqueSettingReq,
	data *GetUniqueSettingData,
) (*GetUniqueSettingResp, error) {
	db := uc.DB

	err := uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type, false,
		data.ExtraLoadOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if setting != nil {
		setting.CurrentObjectID = req.Scope.MainObjectID()
	}

	refObjects, err := uc.SettingService.LoadReferenceObjects(ctx, db, req.Scope, true, false, setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &GetUniqueSettingResp{
		Data:       setting,
		RefObjects: refObjects,
	}, nil
}
