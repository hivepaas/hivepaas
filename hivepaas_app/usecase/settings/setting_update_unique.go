package settings

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

type UpdateUniqueSettingReq struct {
	BaseSettingReq
	AvailInProjects bool `json:"availableInProjects"`
	Default         bool `json:"default"`
	UpdateVer       int  `json:"updateVer"`
}

type UpdateUniqueSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type UpdateUniqueSettingData struct {
	BaseSettingData

	Setting *entity.Setting

	Name            string
	Kind            string
	Version         int
	VerifyingRefIDs []string
	ExtraLoadOpts   []bunex.SelectQueryOption

	Load             func(context.Context, database.Tx, *UpdateUniqueSettingData) error
	AfterLoading     func(context.Context, database.Tx, *UpdateUniqueSettingData) error
	PrepareUpdate    func(context.Context, database.Tx, *UpdateUniqueSettingData, *PersistingSettingData) error
	BeforePersisting func(context.Context, database.Tx, *UpdateUniqueSettingData, *PersistingSettingData) error
	AfterPersisting  func(context.Context, database.Tx, *UpdateUniqueSettingData, *PersistingSettingData) error
}

func (uc *BaseUC) UpdateUniqueSetting(
	ctx context.Context,
	req *UpdateUniqueSettingReq,
	data *UpdateUniqueSettingData,
) (*UpdateUniqueSettingResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadUniqueSettingForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingData{}
		uc.prepareUniqueSettingUpdate(req, data, persistingData)

		if data.PrepareUpdate != nil {
			if err := data.PrepareUpdate(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		err = uc.persistUniqueSettingUpdate(ctx, db, req, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		// Fire update event
		err = uc.SettingService.OnUpdate(ctx, db, &settingservice.UpdateEvent{
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

	return &UpdateUniqueSettingResp{}, nil
}

func (uc *BaseUC) loadUniqueSettingForUpdate(
	ctx context.Context,
	db database.Tx,
	req *UpdateUniqueSettingReq,
	data *UpdateUniqueSettingData,
) (err error) {
	err = uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.Load != nil {
		err = data.Load(ctx, db, data)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
	} else {
		loadOpts := []bunex.SelectQueryOption{
			bunex.SelectFor("UPDATE OF setting"),
		}
		loadOpts = append(loadOpts, data.ExtraLoadOpts...)
		setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type, false, loadOpts...)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		data.Setting = setting
	}

	// Not allow updating inherited settings, in the case, create a new one overriding the upstream
	if data.Setting != nil && data.Setting.ObjectID != req.Scope.MainObjectID() {
		data.Setting = nil
	}

	if data.Setting != nil && req.UpdateVer != data.Setting.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	// Verify that the referenced settings exist
	if len(data.VerifyingRefIDs) > 0 {
		err := uc.checkRefSettingsExistence(ctx, db, &req.BaseSettingReq, data.VerifyingRefIDs, true)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *BaseUC) prepareUniqueSettingUpdate(
	req *UpdateUniqueSettingReq,
	data *UpdateUniqueSettingData,
	persistingData *PersistingSettingData,
) {
	timeNow := timeutil.NowUTC()
	var setting *entity.Setting
	if data.Setting == nil {
		setting = &entity.Setting{
			ID:              gofn.Must(ulid.NewStringULID()),
			Scope:           req.Scope.ScopeType(),
			ObjectID:        req.Scope.MainObjectID(),
			Type:            req.Type,
			Status:          base.SettingStatusActive,
			Name:            data.Name,
			Kind:            data.Kind,
			AvailInProjects: gofn.If(!req.Scope.IsGlobalScope(), false, req.AvailInProjects),
			Default:         req.Default,
			Version:         data.Version,
			CreatedAt:       timeNow,
			UpdatedAt:       timeNow,
		}
	} else {
		copySetting := *data.Setting
		setting = &copySetting
		setting.Name = gofn.Coalesce(data.Name, setting.Name)
		setting.AvailInProjects = gofn.If(!req.Scope.IsGlobalScope(), false, req.AvailInProjects)
		setting.Default = req.Default
		setting.UpdateVer++
		setting.UpdatedAt = timeNow
	}

	persistingData.Setting = setting
}

func (uc *BaseUC) persistUniqueSettingUpdate(
	ctx context.Context,
	db database.IDB,
	req *UpdateUniqueSettingReq,
	persistingData *PersistingSettingData,
) error {
	err := uc.SettingRepo.Upsert(ctx, db, persistingData.Setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.SettingRepo.EnsureUnique(ctx, db, req.Scope, req.Type)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
