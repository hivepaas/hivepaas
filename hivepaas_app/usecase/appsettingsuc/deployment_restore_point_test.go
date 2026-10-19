package appsettingsuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

// appsByID answers the apps it holds, by id, and not found for any other.
type appsByID struct {
	appservice.Service
	apps map[string]*entity.App
}

func (f *appsByID) LoadApp(_ context.Context, _ database.IDB, _, appID string, _, _ bool,
	_ ...bunex.SelectQueryOption) (*entity.App, error) {
	if app, found := f.apps[appID]; found {
		return app, nil
	}
	return nil, hperrors.NewNotFound("App")
}

func managedDatabase(id, env string, managed *entity.AppManagedSettings) *entity.App {
	app := &entity.App{ID: id, ProjectEnvID: env}
	if managed != nil {
		setting := &entity.Setting{ID: id + "-managed", Type: base.SettingTypeAppManaged, Status: base.SettingStatusActive}
		setting.MustSetData(managed)
		app.Settings = []*entity.Setting{setting}
	}
	return app
}

// A deployment can mark a point only on a managed database of its own
// environment, other than itself, whatever its engine; not on a replica, whose
// archive is its primary's.
func TestTheDatabaseAnAppMarksPointsOn(t *testing.T) {
	postgres := &entity.AppManagedSettings{Engine: "postgres", Database: &entity.ManagedDatabase{}}
	mariadb := &entity.AppManagedSettings{Engine: "mariadb", Database: &entity.ManagedDatabase{}}
	web := &entity.App{ID: "web", ProjectID: "p1", ProjectEnvID: "dev"}
	uc := &UC{appService: &appsByID{apps: map[string]*entity.App{
		"db":        managedDatabase("db", "dev", postgres),
		"other-db":  managedDatabase("other-db", "dev", mariadb),
		"other-env": managedDatabase("other-env", "prod", postgres),
		"plain":     managedDatabase("plain", "dev", nil),
		"replica": managedDatabase("replica", "dev", &entity.AppManagedSettings{Engine: "postgres",
			MemberOf: &entity.ManagedMemberOf{App: entity.ObjectID{ID: "db"}, Role: entity.ManagedRoleReplica}}),
	}}}

	assert.NoError(t, uc.checkRestorePointDatabase(context.Background(), nil, web, ""), "none asked for")
	assert.NoError(t, uc.checkRestorePointDatabase(context.Background(), nil, web, "db"))
	assert.NoError(t, uc.checkRestorePointDatabase(context.Background(), nil, web, "other-db"), "any engine's")
	for id, why := range map[string]string{
		"web":       "itself",
		"gone":      "no such app",
		"other-env": "not in the app's environment",
		"plain":     "no engine looks after",
		"replica":   "not on a replica",
	} {
		err := uc.checkRestorePointDatabase(context.Background(), nil, web, id)
		assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid, id)
		info, _ := hperrors.ParseError(err, "en")
		assert.Contains(t, info.DebugLog+info.Detail, why, id)
	}
}
