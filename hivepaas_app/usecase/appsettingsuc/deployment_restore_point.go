package appsettingsuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// restorePointField is the deployment settings' field naming the database.
const restorePointField = "restorePointBeforeDeploy"

// checkRestorePointDatabase refuses a database a deployment of app cannot mark
// a restore point on: the app itself, an app outside its environment, or one no
// PostgreSQL engine looks after.
func (uc *UC) checkRestorePointDatabase(ctx context.Context, db database.IDB, app *entity.App,
	databaseID string) error {
	if databaseID == "" {
		return nil
	}
	invalid := func(why string) error {
		return hperrors.NewArgumentInvalid(restorePointField).WithExtraDetail("%s", why)
	}
	if databaseID == app.ID {
		return invalid("an app cannot mark a restore point on itself")
	}
	target, err := uc.appService.LoadApp(ctx, db, app.ProjectID, databaseID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),
	)
	if errors.Is(err, hperrors.ErrNotFound) {
		return invalid("no such app in the project")
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	if target.ProjectEnvID != app.ProjectEnvID {
		return invalid("the database is not in the app's environment")
	}
	setting := target.GetSettingByType(base.SettingTypeAppManaged)
	if setting == nil {
		return invalid("no engine looks after the database")
	}
	managed, err := setting.AsAppManagedSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if managed.Postgres == nil {
		return invalid("restore points are made on a managed PostgreSQL database")
	}
	return nil
}
