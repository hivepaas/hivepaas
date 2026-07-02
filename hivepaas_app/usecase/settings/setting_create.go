package settings

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

type CreateSettingReq struct {
	BaseSettingReq
	AvailInProjects bool `json:"availableInProjects"`
	Default         bool `json:"default"`
}

type CreateSettingResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}

type CreateSettingData struct {
	BaseSettingData

	VerifyingName       string
	VerifyingRefIDs     *entity.RefObjectIDs
	MultiDefaultAllowed bool
	Version             int

	AfterLoading     func(context.Context, database.Tx, *CreateSettingData) error
	PrepareCreation  func(context.Context, database.Tx, *CreateSettingData, *PersistingSettingCreationData) error
	BeforePersisting func(context.Context, database.Tx, *CreateSettingData, *PersistingSettingCreationData) error
	AfterPersisting  func(context.Context, database.Tx, *CreateSettingData, *PersistingSettingCreationData) error
}

type PersistingSettingCreationData struct {
	Setting *entity.Setting
}

func (uc *BaseUC) CreateSetting(
	ctx context.Context,
	req *CreateSettingReq,
	data *CreateSettingData,
) (*CreateSettingResp, error) {
	var persistingData *PersistingSettingCreationData
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForCreation(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.New(err)
			}
		}

		persistingData = &PersistingSettingCreationData{}
		uc.prepareSettingCreation(req, data, persistingData)

		if data.PrepareCreation != nil {
			if err := data.PrepareCreation(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		err = uc.persistSettingCreation(ctx, db, req, data, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		// Fire create event
		err = uc.SettingService.OnCreate(ctx, db, &settingservice.CreateEvent{Setting: persistingData.Setting})
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &CreateSettingResp{
		Data: &basedto.ObjectIDResp{ID: persistingData.Setting.ID},
	}, nil
}

func (uc *BaseUC) loadSettingForCreation(
	ctx context.Context,
	db database.IDB,
	req *CreateSettingReq,
	data *CreateSettingData,
) error {
	err := uc.loadSettingScopeData(ctx, db, &req.BaseSettingReq, &data.BaseSettingData)
	if err != nil {
		return apperrors.New(err)
	}

	// Verify that the name is available to use
	if data.VerifyingName != "" {
		err := uc.checkNameConflict(ctx, db, &req.BaseSettingReq, data.VerifyingName)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Verify that the referenced objects exist
	err = uc.checkRefObjectsExistence(ctx, db, &req.BaseSettingReq, data.VerifyingRefIDs, true)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (uc *BaseUC) prepareSettingCreation(
	req *CreateSettingReq,
	data *CreateSettingData,
	persistingData *PersistingSettingCreationData,
) {
	timeNow := timeutil.NowUTC()
	setting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           req.Scope.ScopeType(),
		ObjectID:        req.Scope.MainObjectID(),
		Type:            req.Type,
		Status:          base.SettingStatusActive,
		Name:            data.VerifyingName,
		AvailInProjects: gofn.If(!req.Scope.IsGlobalScope(), false, req.AvailInProjects),
		Default:         req.Default,
		Version:         data.Version,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	persistingData.Setting = setting
}

func (uc *BaseUC) persistSettingCreation(
	ctx context.Context,
	db database.IDB,
	req *CreateSettingReq,
	data *CreateSettingData,
	persistingData *PersistingSettingCreationData,
) error {
	err := uc.SettingRepo.Upsert(ctx, db, persistingData.Setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	if !data.MultiDefaultAllowed && persistingData.Setting.Default {
		err = uc.ensureSettingDefaultUniqueness(ctx, db, &req.BaseSettingReq, persistingData.Setting)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}
