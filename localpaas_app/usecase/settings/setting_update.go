package settings

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
)

type UpdateSettingReq struct {
	BaseSettingReq
	ID              string `json:"-"`
	AvailInProjects bool   `json:"availableInProjects"`
	Default         bool   `json:"default"`
	UpdateVer       int    `json:"updateVer"`
}

type UpdateSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type UpdateSettingData struct {
	Setting *entity.Setting

	VerifyingName     string
	VerifyingRefIDs   []string
	DefaultMustUnique bool
	ExtraLoadOpts     []bunex.SelectQueryOption

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
			return apperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return apperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingData{}
		uc.prepareSettingUpdate(req, data, persistingData)

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

		err = uc.persistSettingUpdate(ctx, db, req, data, persistingData)
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

	return &UpdateSettingResp{}, nil
}

func (uc *BaseUC) loadSettingForUpdate(
	ctx context.Context,
	db database.Tx,
	req *UpdateSettingReq,
	data *UpdateSettingData,
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

	// If name changes, validate the new one
	if data.VerifyingName != "" && !strings.EqualFold(setting.Name, data.VerifyingName) {
		err = uc.checkNameConflict(ctx, db, &req.BaseSettingReq, data.VerifyingName)
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
