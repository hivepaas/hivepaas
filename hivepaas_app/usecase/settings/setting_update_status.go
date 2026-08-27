package settings

import (
	"context"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
)

type UpdateSettingStatusReq struct {
	BaseSettingReq
	ID          string              `json:"-"`
	Status      *base.SettingStatus `json:"status"`
	ExpireAt    *time.Time          `json:"expireAt"`
	Inheritable *bool               `json:"inheritable"`
	Default     *bool               `json:"default"`
	UpdateVer   int                 `json:"updateVer"`
}

func (req *UpdateSettingStatusReq) Validate() (validators []vld.Validator) {
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return
}

type UpdateSettingStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type UpdateSettingStatusData struct {
	BaseSettingData

	Setting *entity.Setting

	MultiDefaultAllowed bool
	ExtraLoadOpts       []bunex.SelectQueryOption

	Load             func(context.Context, database.Tx, *UpdateSettingStatusData) error
	AfterLoading     func(context.Context, database.Tx, *UpdateSettingStatusData) error
	BeforePersisting func(context.Context, database.Tx, *UpdateSettingStatusData, *PersistingSettingStatusData) error
	AfterPersisting  func(context.Context, database.Tx, *UpdateSettingStatusData, *PersistingSettingStatusData) error
}

type PersistingSettingStatusData struct {
	Setting *entity.Setting
}

func (uc *BaseUC) UpdateSettingStatus(
	ctx context.Context,
	req *UpdateSettingStatusReq,
	data *UpdateSettingStatusData,
) (*UpdateSettingStatusResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForUpdateStatus(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return hperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingStatusData{}
		uc.prepareSettingStatusUpdate(req, data, persistingData)
		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return hperrors.Wrap(err)
			}
		}

		err = uc.persistSettingStatusUpdate(ctx, db, req, data, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return hperrors.Wrap(err)
			}
		}

		// Fire update event
		err = uc.SettingEventService.OnUpdateStatus(ctx, db, &settingeventservice.UpdateEvent{
			Setting:    persistingData.Setting,
			OldSetting: data.Setting,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &UpdateSettingStatusResp{}, nil
}

func (uc *BaseUC) loadSettingForUpdateStatus(
	ctx context.Context,
	db database.Tx,
	req *UpdateSettingStatusReq,
	data *UpdateSettingStatusData,
) (err error) {
	if err = uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return hperrors.Wrap(err)
	}

	if data.Load != nil {
		err = data.Load(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
	} else {
		loadOpts := []bunex.SelectQueryOption{
			bunex.SelectFor("UPDATE OF setting"),
		}
		loadOpts = append(loadOpts, data.ExtraLoadOpts...)

		setting, err := uc.loadSettingByID(ctx, db, &req.BaseSettingReq, req.ID,
			false, loadOpts...)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.Setting = setting
	}

	setting := data.Setting
	if req.UpdateVer != setting.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	if setting.ObjectID != req.Scope.ScopeObjectID() {
		return hperrors.Wrap(hperrors.ErrInheritedSettingNonUpdatable)
	}

	return nil
}

func (uc *BaseUC) prepareSettingStatusUpdate(
	req *UpdateSettingStatusReq,
	data *UpdateSettingStatusData,
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
	if req.Inheritable != nil {
		setting.Inheritable = *req.Inheritable
	}
	if req.Default != nil {
		setting.Default = *req.Default
	}

	persistingData.Setting = setting
}

func (uc *BaseUC) persistSettingStatusUpdate(
	ctx context.Context,
	db database.IDB,
	req *UpdateSettingStatusReq,
	data *UpdateSettingStatusData,
	persistingData *PersistingSettingStatusData,
) error {
	err := uc.SettingRepo.Update(ctx, db, persistingData.Setting)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if !data.MultiDefaultAllowed && !data.Setting.Default && persistingData.Setting.Default {
		err = uc.ensureSettingDefaultUniqueness(ctx, db, &req.BaseSettingReq, persistingData.Setting)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}
