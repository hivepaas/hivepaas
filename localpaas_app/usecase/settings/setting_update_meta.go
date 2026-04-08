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

type UpdateSettingMetaReq struct {
	BaseSettingReq
	ID                  string              `json:"-"`
	Status              *base.SettingStatus `json:"status"`
	ExpireAt            *time.Time          `json:"expireAt"`
	AvailableInProjects *bool               `json:"availableInProjects"`
	Default             *bool               `json:"default"`
	UpdateVer           int                 `json:"updateVer"`
}

type UpdateSettingMetaResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type UpdateSettingMetaData struct {
	Setting *entity.Setting

	DefaultMustUnique bool
	ExtraLoadOpts     []bunex.SelectQueryOption

	Load             func(context.Context, database.Tx, *UpdateSettingMetaData) error
	AfterLoading     func(context.Context, database.Tx, *UpdateSettingMetaData) error
	BeforePersisting func(context.Context, database.Tx, *UpdateSettingMetaData, *PersistingSettingMetaData) error
	AfterPersisting  func(context.Context, database.Tx, *UpdateSettingMetaData, *PersistingSettingMetaData) error
}

type PersistingSettingMetaData struct {
	Setting *entity.Setting
}

func (uc *BaseUC) UpdateSettingMeta(
	ctx context.Context,
	req *UpdateSettingMetaReq,
	data *UpdateSettingMetaData,
) (*UpdateSettingMetaResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForUpdateMeta(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingMetaData{}
		uc.prepareSettingMetaUpdate(req, data, persistingData)
		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		err = uc.persistSettingMetaUpdate(ctx, db, req, data, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		// Fire update event
		err = uc.SettingService.OnUpdateMeta(ctx, db, &settingservice.UpdateEvent{
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

	return &UpdateSettingMetaResp{}, nil
}

func (uc *BaseUC) loadSettingForUpdateMeta(
	ctx context.Context,
	db database.Tx,
	req *UpdateSettingMetaReq,
	data *UpdateSettingMetaData,
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

		setting, err := uc.loadSettingByID(ctx, db, &req.BaseSettingReq, req.ID,
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

func (uc *BaseUC) prepareSettingMetaUpdate(
	req *UpdateSettingMetaReq,
	data *UpdateSettingMetaData,
	persistingData *PersistingSettingMetaData,
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

func (uc *BaseUC) persistSettingMetaUpdate(
	ctx context.Context,
	db database.IDB,
	req *UpdateSettingMetaReq,
	data *UpdateSettingMetaData,
	persistingData *PersistingSettingMetaData,
) error {
	err := uc.SettingRepo.Update(ctx, db, persistingData.Setting,
		bunex.UpdateColumns("update_ver", "updated_at", "status", "expire_at", "avail_in_projects", "is_default"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.DefaultMustUnique && !data.Setting.Default && persistingData.Setting.Default {
		if data.DefaultMustUnique && persistingData.Setting.Default {
			err = uc.ensureSettingDefaultUniqueness(ctx, db, &req.BaseSettingReq, persistingData.Setting)
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	return nil
}
