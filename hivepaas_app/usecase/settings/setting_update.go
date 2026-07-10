package settings

import (
	"context"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

type UpdateSettingReq struct {
	BaseSettingReq
	ID              string `json:"-"`
	AvailInProjects bool   `json:"availableInProjects"`
	Default         bool   `json:"default"`
	UpdateVer       int    `json:"updateVer"`
}

func (req *UpdateSettingReq) Validate() (validators []vld.Validator) {
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return
}

type UpdateSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type UpdateSettingData struct {
	BaseSettingData

	Setting *entity.Setting

	VerifyingName       string
	VerifyingRefIDs     *entity.RefObjectIDs
	MultiDefaultAllowed bool
	ExtraLoadOpts       []bunex.SelectQueryOption

	Load             func(context.Context, database.Tx, *UpdateSettingData) error
	AfterLoading     func(context.Context, database.Tx, *UpdateSettingData) error
	PrepareUpdate    func(context.Context, database.Tx, *UpdateSettingData, *PersistingSettingData) error
	BeforePersisting func(context.Context, database.Tx, *UpdateSettingData, *PersistingSettingData) error
	AfterPersisting  func(context.Context, database.Tx, *UpdateSettingData, *PersistingSettingData) error
}

type PersistingSettingData struct {
	Setting *entity.Setting
}

func (uc *BaseUC) UpdateSetting(
	ctx context.Context,
	req *UpdateSettingReq,
	data *UpdateSettingData,
) (*UpdateSettingResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.New(err)
			}
		}

		persistingData := &PersistingSettingData{}
		uc.prepareSettingUpdate(req, data, persistingData)

		if data.PrepareUpdate != nil {
			if err := data.PrepareUpdate(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		err = uc.persistSettingUpdate(ctx, db, req, data, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return apperrors.New(err)
			}
		}

		// Fire update event
		err = uc.SettingService.OnUpdate(ctx, db, &settingservice.UpdateEvent{
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

	return &UpdateSettingResp{}, nil
}

func (uc *BaseUC) loadSettingForUpdate(
	ctx context.Context,
	db database.Tx,
	req *UpdateSettingReq,
	data *UpdateSettingData,
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
		setting, err := uc.loadSettingByID(ctx, db, &req.BaseSettingReq, req.ID,
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

	// If name changes, validate the new one
	if data.VerifyingName != "" && !strings.EqualFold(setting.Name, data.VerifyingName) {
		err = uc.checkNameConflict(ctx, db, &req.BaseSettingReq, data.VerifyingName)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Verify that the referenced settings exist
	err = uc.checkRefObjectsExistence(ctx, db, &req.BaseSettingReq, data.VerifyingRefIDs, true)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (uc *BaseUC) prepareSettingUpdate(
	req *UpdateSettingReq,
	data *UpdateSettingData,
	persistingData *PersistingSettingData,
) {
	timeNow := timeutil.NowUTC()
	copySetting := *data.Setting
	setting := &copySetting
	setting.Name = gofn.Coalesce(data.VerifyingName, setting.Name)
	setting.AvailInProjects = gofn.If(!req.Scope.IsGlobalScope(), false, req.AvailInProjects)
	setting.Default = req.Default
	setting.UpdateVer++
	setting.UpdatedAt = timeNow

	persistingData.Setting = setting
}

func (uc *BaseUC) persistSettingUpdate(
	ctx context.Context,
	db database.IDB,
	req *UpdateSettingReq,
	data *UpdateSettingData,
	persistingData *PersistingSettingData,
) error {
	err := uc.SettingRepo.Update(ctx, db, persistingData.Setting)
	if err != nil {
		return apperrors.New(err)
	}

	if !data.MultiDefaultAllowed && !data.Setting.Default && persistingData.Setting.Default {
		err = uc.ensureSettingDefaultUniqueness(ctx, db, &req.BaseSettingReq, persistingData.Setting)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}
