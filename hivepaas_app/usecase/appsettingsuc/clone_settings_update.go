package appsettingsuc

import (
	"context"
	"time"

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppCloneSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppCloneSettingsReq,
) (*appsettingsdto.UpdateAppCloneSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppCloneSettingsData{}
		err := uc.loadAppCloneSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppCloneSettings(ctx, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppCloneSettingsResp{
		Meta: &basedto.Meta{},
	}, nil
}

type updateAppCloneSettingsData struct {
	App                 *entity.App
	AppCloneSetting     *entity.Setting
	NewAppCloneSettings *entity.AppCloneSettings
	RefObjects          *entity.RefObjects
}

func (uc *UC) loadAppCloneSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppCloneSettingsReq,
	data *updateAppCloneSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppClone),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app
	data.AppCloneSetting = app.GetSettingByType(base.SettingTypeAppClone)

	if data.AppCloneSetting != nil && data.AppCloneSetting.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	newAppCloneSettings := req.ToEntity()
	data.NewAppCloneSettings = newAppCloneSettings

	data.RefObjects, err = uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetObjectScope(),
		true, true, newAppCloneSettings.GetRefObjectIDs())
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppCloneSettings(
	_ context.Context,
	data *updateAppCloneSettingsData,
	persistingData *persistingAppData,
) {
	app := data.App
	setting := data.AppCloneSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppClone,
			CreatedAt: timeNow,
			Version:   entity.CurrentAppCloneSettingsVersion,
		}
		data.AppCloneSetting = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.Status = base.SettingStatusActive
	setting.ExpireAt = time.Time{}
	setting.MustSetData(data.NewAppCloneSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}
