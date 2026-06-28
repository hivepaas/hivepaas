package appservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
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

	ExecuteInTx(ctx context.Context, app *entity.App, requireUpdateVerMatch bool, fn func(database.Tx) error) error
}
