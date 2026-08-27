package appsettingsuc

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) ExecuteAppClone(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.ExecuteAppCloneReq,
) (*appsettingsdto.ExecuteAppCloneResp, error) {
	var data *executeAppCloneData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &executeAppCloneData{}
		err := uc.loadAppCloneSettingsForExecute(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, data.AppCloneTask)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if data != nil && data.AppCloneTask != nil {
		if err = uc.taskQueue.ScheduleTask(ctx, data.AppCloneTask); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return &appsettingsdto.ExecuteAppCloneResp{}, nil
}

type executeAppCloneData struct {
	App          *entity.App
	AppCloneTask *entity.Task
}

func (uc *UC) loadAppCloneSettingsForExecute(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.ExecuteAppCloneReq,
	data *executeAppCloneData,
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
		return hperrors.Wrap(err)
	}
	data.App = app
	cloneSetting := app.GetSettingByType(base.SettingTypeAppClone)

	if cloneSetting == nil {
		return hperrors.NewNotFound("App clone settings")
	}
	cloneSettings := cloneSetting.MustAsAppCloneSettings()

	// Validate target app's name to be available
	appKey := projecthelper.CalcAppKey(cloneSettings.TargetName)
	targetEnv := gofn.Coalesce(cloneSettings.TargetEnv, app.ProjectEnv.Key)
	appGlobalKey := projecthelper.CalcAppGlobalKey(app.Project.Key, appKey, targetEnv)
	// App keys must be unique globally
	conflictApp, err := uc.appRepo.GetByGlobalKey(ctx, db, "", appGlobalKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if conflictApp != nil {
		return hperrors.NewAlreadyExist("App").
			WithMsgLog("app unique key '%s' already exists", appGlobalKey)
	}

	// Create a task for cloning the app
	cloneTask, err := uc.appCloneService.CreateAppCloneTask(data.App)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.AppCloneTask = cloneTask

	return nil
}
