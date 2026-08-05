package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) ExecuteAppClone(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.ExecuteAppCloneReq,
) (*appsettingsdto.ExecuteAppCloneResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &executeAppCloneData{}
		err := uc.loadAppCloneSettingsForExecute(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// TODO: execute the clone by creating a task

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.ExecuteAppCloneResp{}, nil
}

type executeAppCloneData struct {
	App              *entity.App
	AppCloneSetting  *entity.Setting
	AppCloneSettings *entity.AppCloneSettings
	RefObjects       *entity.RefObjects
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
		return apperrors.Wrap(err)
	}
	data.App = app
	cloneSetting := app.GetSettingByType(base.SettingTypeAppClone)

	if cloneSetting == nil {
		return apperrors.NewNotFound("App clone settings")
	}
	data.AppCloneSetting = cloneSetting
	data.AppCloneSettings = cloneSetting.MustAsAppCloneSettings()

	// TODO: Validate new app's name

	return nil
}
