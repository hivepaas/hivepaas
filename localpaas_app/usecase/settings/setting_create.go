package settings

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
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
	VerifyingName     string
	VerifyingRefIDs   []string
	DefaultMustUnique bool
	Version           int

	AfterLoading     func(context.Context, database.Tx, *CreateSettingData) error
	PrepareCreation  func(context.Context, database.Tx, *CreateSettingData, *PersistingSettingCreationData) error
	BeforePersisting func(context.Context, database.Tx, *CreateSettingData, *PersistingSettingCreationData) error
	AfterPersisting  func(context.Context, database.Tx, *CreateSettingData, *PersistingSettingCreationData) error
}

type PersistingSettingCreationData struct {
	Setting *entity.Setting
}

func (uc *BaseSettingUC) CreateSetting(
	ctx context.Context,
	req *CreateSettingReq,
	data *CreateSettingData,
) (*CreateSettingResp, error) {
	var persistingData *PersistingSettingCreationData
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForCreation(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData = &PersistingSettingCreationData{}
		uc.prepareSettingCreation(req, data, persistingData)

		if data.PrepareCreation != nil {
			if err := data.PrepareCreation(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		err = uc.persistSettingCreation(ctx, db, req, data, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.Wrap(err)
			}
		}

		// Fire create event
		err = uc.SettingService.OnCreate(ctx, db, &settingservice.CreateEvent{Setting: persistingData.Setting})
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &CreateSettingResp{
		Data: &basedto.ObjectIDResp{ID: persistingData.Setting.ID},
	}, nil
}

func (uc *BaseSettingUC) loadSettingForCreation(
	ctx context.Context,
	db database.IDB,
	req *CreateSettingReq,
	data *CreateSettingData,
) error {
	// Verify that the name is available to use
	if data.VerifyingName != "" {
		err := uc.checkNameConflict(ctx, db, &req.BaseSettingReq, data.VerifyingName)
		if err != nil {
			return apperrors.Wrap(err)
		}
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

func (uc *BaseSettingUC) prepareSettingCreation(
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

func (uc *BaseSettingUC) persistSettingCreation(
	ctx context.Context,
	db database.IDB,
	req *CreateSettingReq,
	data *CreateSettingData,
	persistingData *PersistingSettingCreationData,
) error {
	err := uc.SettingRepo.Upsert(ctx, db, persistingData.Setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.DefaultMustUnique && persistingData.Setting.Default {
		err = uc.ensureSettingDefaultUniqueness(ctx, db, &req.BaseSettingReq, persistingData.Setting)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
