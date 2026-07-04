package appservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	LoadApps(ctx context.Context, db database.IDB, projectID string, appIDs []string,
		requireProjectActive, requireAppsActive bool, extraOpts ...bunex.SelectQueryOption) (
		[]*entity.App, error)
	LoadApp(ctx context.Context, db database.IDB, projectID, appID string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)
	LoadAppByKey(ctx context.Context, db database.IDB, projectID, appKey string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)
	LoadAppWithFeatureSettings(ctx context.Context, db database.IDB, projectID, appID string,
		requireProjectActive, requireAppActive bool, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, *entity.AppFeatureSettings, error)

	FindAppsMatchingRepository(ctx context.Context, db database.IDB, repoID, repoRef string,
		extraAppOpts ...bunex.SelectQueryOption) ([]*entity.App, error)

	PersistAppData(ctx context.Context, db database.IDB, data *PersistingAppData) error
	DeleteApp(ctx context.Context, db database.IDB, app *entity.App) error
	SetAppStatus(ctx context.Context, db database.IDB, app *entity.App, status base.AppStatus, recursive bool) error
	SetAppRunning(ctx context.Context, app *entity.App, running bool) error

	ExecuteInTx(ctx context.Context, app *entity.App, requireUpdateVerMatch bool, fn func(database.Tx) error) error
}
