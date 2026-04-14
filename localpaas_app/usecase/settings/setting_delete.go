package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
)

type DeleteSettingReq struct {
	BaseSettingReq
	ID string `json:"-" mapstructure:"-"`
}

func (req *DeleteSettingReq) Validate() (validators []vld.Validator) {
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return
}

type DeleteSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type DeleteSettingData struct {
	BaseSettingData

	Setting              *entity.Setting
	ProjectSharedSetting *entity.ProjectSharedSetting
	ExtraLoadOpts        []bunex.SelectQueryOption

	AfterLoading     func(context.Context, database.Tx, *DeleteSettingData) error
	BeforePersisting func(context.Context, database.Tx, *DeleteSettingData, *PersistingSettingDeletionData) error
	AfterPersisting  func(context.Context, database.Tx, *DeleteSettingData, *PersistingSettingDeletionData) error
}

type PersistingSettingDeletionData struct {
	Setting              *entity.Setting
	ProjectSharedSetting *entity.ProjectSharedSetting
}

func (uc *BaseUC) DeleteSetting(
	ctx context.Context,
	req *DeleteSettingReq,
	data *DeleteSettingData,
) (*DeleteSettingResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForDeletion(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingDeletionData{}
		uc.prepareSettingDeletion(req, data, persistingData)
		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		err = uc.persistSettingDeletion(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		// Fire delete event
		err = uc.SettingService.OnDelete(ctx, db, &settingservice.DeleteEvent{Setting: persistingData.Setting})
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &DeleteSettingResp{}, nil
}

func (uc *BaseUC) loadSettingForDeletion(
	ctx context.Context,
	db database.IDB,
	req *DeleteSettingReq,
	data *DeleteSettingData,
) (err error) {
	err = uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return apperrors.Wrap(err)
	}

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

	// The setting was imported to project from global
	if setting.ObjectID == "" && req.Scope.IsProjectScope() {
		data.ProjectSharedSetting, err = uc.ProjectSharedSettingRepo.Get(ctx, db, req.Scope.ProjectID, req.ID)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *BaseUC) prepareSettingDeletion(
	_ *DeleteSettingReq,
	data *DeleteSettingData,
	persistingData *PersistingSettingDeletionData,
) {
	timeNow := timeutil.NowUTC()
	if data.ProjectSharedSetting != nil {
		data.ProjectSharedSetting.DeletedAt = timeNow
		persistingData.ProjectSharedSetting = data.ProjectSharedSetting
	} else {
		data.Setting.UpdateVer++
		data.Setting.DeletedAt = timeNow
		persistingData.Setting = data.Setting
	}
}

func (uc *BaseUC) persistSettingDeletion(
	ctx context.Context,
	db database.IDB,
	persistingData *PersistingSettingDeletionData,
) (err error) {
	if persistingData.ProjectSharedSetting != nil {
		err = uc.ProjectSharedSettingRepo.Update(ctx, db, persistingData.ProjectSharedSetting,
			bunex.UpdateColumns("deleted_at"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	}

	err = uc.SettingRepo.Update(ctx, db, persistingData.Setting,
		bunex.UpdateColumns("deleted_at"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If deleted item is global, delete all references from projects
	if persistingData.Setting.ObjectID == "" {
		err = uc.ProjectSharedSettingRepo.DeleteAllBySetting(ctx, db, persistingData.Setting.ID)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
