package settings

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
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
		err := uc.loadUniqueSettingForUpdateMeta(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingStatusData{}
		uc.prepareUniqueSettingStatusUpdate(req, data, persistingData)
		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		err = uc.persistUniqueSettingStatusUpdate(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		// Fire update event
		err = uc.SettingService.OnUpdateStatus(ctx, db, &settingservice.UpdateEvent{
			Setting:    persistingData.Setting,
			OldSetting: data.Setting,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &UpdateUniqueSettingStatusResp{}, nil
}

func (uc *BaseUC) loadUniqueSettingForUpdateMeta(
	ctx context.Context,
	db database.Tx,
	req *UpdateUniqueSettingStatusReq,
	data *UpdateUniqueSettingStatusData,
) (err error) {
	if data.Load != nil {
		err = data.Load(ctx, db, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		loadOpts := []bunex.SelectQueryOption{
			bunex.SelectFor("UPDATE OF setting"),
		}
		loadOpts = append(loadOpts, data.ExtraLoadOpts...)

		setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type,
			false, loadOpts...)
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.Setting = setting
	}

	setting := data.Setting
	if req.UpdateVer != setting.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	if setting.ObjectID != req.Scope.MainObjectID() {
		return apperrors.New(apperrors.ErrOwnSettingRequired).
			WithMsgLog("imported or inherited setting is not allowed to update")
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
		return apperrors.Wrap(err)
	}
	return nil
}
