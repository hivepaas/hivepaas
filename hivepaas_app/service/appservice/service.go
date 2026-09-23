package appservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	LoadApps(ctx context.Context, db database.IDB, projectID string, appIDs []string,
		requireProjectActive, requireAppsActive bool, extraOpts ...bunex.SelectQueryOption) (
		[]*entity.App, error)
	LoadAppsSkipMissing(ctx context.Context, db database.IDB, projectID string, appIDs []string,
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
	EnsureAppActive(ctx context.Context, db database.Tx, app *entity.App,
		checkUpdateVer bool, lockApp bool) error

	LoadChildApps(ctx context.Context, db database.IDB, app *entity.App,
		loadChildApps, loadLogicalChildApps bool) error

	FindAppsMatchingRepository(ctx context.Context, db database.IDB, repoID, repoRef string,
		extraAppOpts ...bunex.SelectQueryOption) ([]*entity.App, error)

	PersistAppData(ctx context.Context, db database.IDB, data *PersistingAppData) error
	// DeleteApp removes an app, its service and everything recorded about it.
	// removeStorage also deletes the directories it kept its data in, inside the
	// volumes it mounted; without it those are left where they are.
	DeleteApp(ctx context.Context, db database.IDB, app *entity.App, removeStorage, cascade bool) error
	SetAppStatus(ctx context.Context, db database.IDB, app *entity.App, status base.AppStatus, cascade bool) error
	SetAppRunning(ctx context.Context, app *entity.App, running bool) error
	// RecreateServiceWithSpec deletes and recreates the app swarm service, which is the only way to
	// change its mode variant. It causes downtime and returns the new service ID for the caller to
	// persist. See the implementation for the failure/rollback behavior.
	RecreateServiceWithSpec(ctx context.Context, app *entity.App,
		oldSpec, newSpec *swarm.ServiceSpec) (string, error)

	RevealSecrets(ctx context.Context, db database.IDB, auth *basedto.Auth, app *entity.App,
		setting *entity.Setting) (revealed bool, err error)

	ExecuteInTx(ctx context.Context, app *entity.App, requireUpdateVerMatch bool, fn func(database.Tx) error) error
}
