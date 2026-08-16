package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type ListSettingReq struct {
	BaseSettingReq
	Statuses []base.SettingStatus `json:"-" mapstructure:"status"`
	Kinds    []string             `json:"-" mapstructure:"kind"`
	Search   string               `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func (req *ListSettingReq) Validate() (validators []vld.Validator) {
	validators = append(validators, basedto.ValidateSlice(req.Statuses, true, 0,
		base.AllSettingStatuses, "status")...)
	return
}

type ListSettingResp struct {
	Meta       *basedto.ListMeta
	Data       []*entity.Setting
	RefObjects *entity.RefObjects
}

type ListSettingData struct {
	BaseSettingData

	SkipLoadingRefObjects bool
	ExtraLoadOpts         []bunex.SelectQueryOption

	AfterLoading func(context.Context, database.IDB, *ListSettingData) error
}

func (uc *BaseUC) ListSetting(
	ctx context.Context,
	auth *basedto.Auth,
	req *ListSettingReq,
	data *ListSettingData,
) (_ *ListSettingResp, err error) {
	db := uc.DB

	if err = uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return nil, apperrors.Wrap(err)
	}

	if data.AfterLoading != nil {
		if err = data.AfterLoading(ctx, db, data); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.type = ?", req.Type),
	}
	if len(req.Statuses) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("setting.status IN (?)", req.Statuses...))
	}
	if len(req.Kinds) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("setting.kind IN (?)", req.Kinds...))
	}
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("setting.name ILIKE ?", keyword),
			),
		)
	}
	allowedAllIDs, allowedIDs := auth.AllowedSettings(nil)
	if !allowedAllIDs {
		if len(allowedIDs) == 0 { // return empty result
			return &ListSettingResp{Meta: basedto.NewEmptyListMeta()}, nil
		}
		listOpts = append(listOpts,
			bunex.SelectWhereIn("setting.id IN (?)", allowedIDs...),
		)
	}
	listOpts = append(listOpts, data.ExtraLoadOpts...)

	settings, paging, err := uc.SettingRepo.List(ctx, db, req.Scope, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, setting := range settings {
		setting.CurrentObjectID = req.Scope.ScopeObjectID()
	}

	var refObjects *entity.RefObjects
	if !data.SkipLoadingRefObjects {
		err = uc.SettingService.LoadRefObjectsSkipMissing(ctx, db, &refObjects, req.Scope, false, settings...)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &ListSettingResp{
		Meta:       &basedto.ListMeta{Page: paging},
		Data:       settings,
		RefObjects: refObjects,
	}, nil
}
