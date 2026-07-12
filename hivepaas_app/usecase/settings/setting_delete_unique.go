package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

type DeleteUniqueSettingReq struct {
	BaseSettingReq
}

func (req *DeleteUniqueSettingReq) Validate() (validators []vld.Validator) {
	return
}

type DeleteUniqueSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type DeleteUniqueSettingData struct {
	BaseSettingData

	Setting              *entity.Setting
	ProjectSharedSetting *entity.ProjectSharedSetting
	ExtraLoadOpts        []bunex.SelectQueryOption

	AfterLoading     func(context.Context, database.Tx, *DeleteUniqueSettingData) error
	BeforePersisting func(context.Context, database.Tx, *DeleteUniqueSettingData, *PersistingSettingDeletionData) error
	AfterPersisting  func(context.Context, database.Tx, *DeleteUniqueSettingData, *PersistingSettingDeletionData) error
}

func (uc *BaseUC) DeleteUniqueSetting(
	ctx context.Context,
	req *DeleteUniqueSettingReq,
	data *DeleteUniqueSettingData,
) (*DeleteUniqueSettingResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadUniqueSettingForDeletion(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingDeletionData{}
		uc.prepareUniqueSettingDeletion(req, data, persistingData)
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

	return &DeleteUniqueSettingResp{}, nil
}

func (uc *BaseUC) loadUniqueSettingForDeletion(
	ctx context.Context,
	db database.IDB,
	req *DeleteUniqueSettingReq,
	data *DeleteUniqueSettingData,
) (err error) {
	err = uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectFor("UPDATE OF setting"),
	}
	loadOpts = append(loadOpts, data.ExtraLoadOpts...)

	setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type, false, loadOpts...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Setting = setting

	// The setting was imported to project from global
	if setting.ObjectID == "" && req.Scope.IsProjectScope() {
		data.ProjectSharedSetting, err = uc.ProjectSharedSettingRepo.Get(ctx, db, req.Scope.ProjectID, setting.ID)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *BaseUC) prepareUniqueSettingDeletion(
	_ *DeleteUniqueSettingReq,
	data *DeleteUniqueSettingData,
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
