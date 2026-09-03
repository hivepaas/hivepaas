package settings

import (
	"context"
	"errors"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
) (_ *GetUniqueSettingResp, err error) {
	db := uc.DB

	if err = uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type, false,
		data.ExtraLoadOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if setting != nil {
		setting.CurrentObjectID = req.Scope.ScopeObjectID()
	}

	refObjects := entity.NewRefObjects()
	err = uc.SettingService.LoadRefObjectsSkipMissing(ctx, db, &refObjects, req.Scope, false, setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &GetUniqueSettingResp{
		Data:       setting,
		RefObjects: refObjects,
	}, nil
}

func (uc *BaseUC) GetUniqueSettingOrEmpty(
	ctx context.Context,
	auth *basedto.Auth,
	req *GetUniqueSettingReq,
	data *GetUniqueSettingData,
) (resp *GetUniqueSettingResp, err error) {
	for i := range 2 { //nolint:mnd
		resp, err = uc.GetUniqueSetting(ctx, auth, req, data)
		if err == nil {
			return resp, nil
		}
		if errors.Is(err, hperrors.ErrNotFound) {
			if i == 0 {
				if e := uc.SettingInitService.InitDefaults(ctx, uc.DB); e != nil {
					return nil, hperrors.Wrap(err)
				}
				continue
			}
			if i == 1 {
				timeNow := timeutil.NowUTC()
				resp = &GetUniqueSettingResp{
					Data: &entity.Setting{
						Scope:     req.Scope.ScopeType,
						ObjectID:  req.Scope.ScopeObjectID(),
						Type:      req.Type,
						Status:    base.SettingStatusActive,
						CreatedAt: timeNow,
						UpdatedAt: timeNow,
					},
					RefObjects: entity.NewRefObjects(),
				}
				return resp, nil
			}
		}
		break
	}
	return nil, hperrors.Wrap(err)
}
