package loggingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/services/logging"
)

type Service interface {
	// Apply makes the cluster match the stored configuration, deploying or removing as needed.
	Apply(ctx context.Context, db database.IDB, req *SettingApplyReq) (*SettingApplyResp, error)

	// TearDown removes the collector and the backend, keeping the data volume.
	TearDown(ctx context.Context) error

	Status(ctx context.Context, db database.IDB, logging *entity.Setting) (*Status, error)

	// QueryAppLogs searches one app's stored logs. The app's identity is put
	// into the query here; nothing in q can widen it.
	QueryAppLogs(ctx context.Context, db database.IDB, app *entity.App, q *AppLogQuery) (*logging.QueryResp, error)

	// AppHistory says whether stored logs can be shown for the app, and if not,
	// why - so the dashboard can say so instead of showing an empty list.
	AppHistory(ctx context.Context, db database.IDB, app *entity.App) (*AppHistory, error)
}
