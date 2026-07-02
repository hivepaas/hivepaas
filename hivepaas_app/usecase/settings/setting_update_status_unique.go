package settings

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

type UpdateUniqueSettingStatusReq struct {
	BaseSettingReq
	Status              *base.SettingStatus `json:"status"`
	ExpireAt            *time.Time          `json:"expireAt"`
	AvailableInProjects *bool               `json:"availableInProjects"`
	Default             *bool               `json:"default"`
	UpdateVer           int                 `json:"updateVer"`
}

type UpdateUniqueSettingStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type UpdateUniqueSettingStatusData struct {
	BaseSettingData

	Setting *entity.Setting

	ExtraLoadOpts []bunex.SelectQueryOption

	Load             func(context.Context, database.Tx, *UpdateUniqueSettingStatusData) error
	AfterLoading     func(context.Context, database.Tx, *UpdateUniqueSettingStatusData) error
	BeforePersisting func(context.Context, database.Tx, *UpdateUniqueSettingStatusData, *PersistingSettingStatusData) error
	AfterPersisting  func(context.Context, database.Tx, *UpdateUniqueSettingStatusData, *PersistingSettingStatusData) error
}

func (uc *BaseUC) UpdateUniqueSettingStatus(
	ctx context.Context,
	req *UpdateUniqueSettingStatusReq,
	data *UpdateUniqueSettingStatusData,
) (*UpdateUniqueSettingStatusResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadUniqueSettingForUpdateStatus(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.New(err)
			}
		}

		persistingData := &PersistingSettingStatusData{}
		uc.prepareUniqueSettingStatusUpdate(req, data, persistingData)
		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		err = uc.persistUniqueSettingStatusUpdate(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		// Fire update event
		err = uc.SettingService.OnUpdateStatus(ctx, db, &settingservice.UpdateEvent{
			Setting:    persistingData.Setting,
			OldSetting: data.Setting,
		})
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &UpdateUniqueSettingStatusResp{}, nil
}

func (uc *BaseUC) loadUniqueSettingForUpdateStatus(
	ctx context.Context,
	db database.Tx,
	req *UpdateUniqueSettingStatusReq,
	data *UpdateUniqueSettingStatusData,
) (err error) {
	err = uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return apperrors.New(err)
	}

	if data.Load != nil {
		err = data.Load(ctx, db, data)
		if err != nil {
			return apperrors.New(err)
		}
	} else {
		loadOpts := []bunex.SelectQueryOption{
			bunex.SelectFor("UPDATE OF setting"),
		}
		loadOpts = append(loadOpts, data.ExtraLoadOpts...)

		setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type,
			false, loadOpts...)
		if err != nil {
			return apperrors.New(err)
		}
		data.Setting = setting
	}

	setting := data.Setting
	if req.UpdateVer != setting.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}

	if setting.ObjectID != req.Scope.MainObjectID() {
		return apperrors.New(apperrors.ErrInheritedSettingNonUpdatable)
	}

	return nil
}

func (uc *BaseUC) prepareUniqueSettingStatusUpdate(
	req *UpdateUniqueSettingStatusReq,
	data *UpdateUniqueSettingStatusData,
	persistingData *PersistingSettingStatusData,
) {
	timeNow := timeutil.NowUTC()
	copySetting := *data.Setting
	setting := &copySetting
	setting.UpdateVer++
	setting.UpdatedAt = timeNow

	if req.Status != nil {
		setting.Status = *req.Status
	}
	if req.ExpireAt != nil {
		setting.ExpireAt = *req.ExpireAt
	}
	if req.AvailableInProjects != nil {
		setting.AvailInProjects = gofn.If(!req.Scope.IsGlobalScope(), false, *req.AvailableInProjects)
	}
	if req.Default != nil {
		setting.Default = *req.Default
	}

	persistingData.Setting = setting
}

func (uc *BaseUC) persistUniqueSettingStatusUpdate(
	ctx context.Context,
	db database.IDB,
	persistingData *PersistingSettingStatusData,
) error {
	err := uc.SettingRepo.Update(ctx, db, persistingData.Setting,
		bunex.UpdateColumns("update_ver", "updated_at", "status", "expire_at", "avail_in_projects", "is_default"),
	)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
