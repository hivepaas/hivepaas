package loggingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/services/logging"
)

type Service interface {
	// Apply makes the cluster match the stored configuration: it provisions the
	// backend and the collector as apps on the first save, reconciles them on the
	// rest, and removes one HivePaaS no longer runs - only when the request
	// confirms it. It is idempotent, so a failed save leaves work the next save
	// retries.
	Apply(ctx context.Context, db database.IDB, req *SettingApplyReq) (*SettingApplyResp, error)

	// Validate refuses a configuration before it is written. The usecase calls it
	// while loading, so a bad save is a validation error rather than a stored
	// configuration Apply then fails on.
	Validate(ctx context.Context, db database.IDB, next, current *entity.LoggingSettings) error

	Status(ctx context.Context, db database.IDB, logging *entity.Setting) (*Status, error)

	// QueryAppLogs searches one app's stored logs. The app's identity is put
	// into the query here; nothing in q can widen it.
	QueryAppLogs(ctx context.Context, db database.IDB, app *entity.App, q *AppLogQuery) (*logging.QueryResp, error)

	// AppHistory says whether stored logs can be shown for the app, and if not,
	// why - so the dashboard can say so instead of showing an empty list.
	AppHistory(ctx context.Context, db database.IDB, app *entity.App) (*AppHistory, error)

	// FunctionMetrics counts a function's calls in its stored logs: one point
	// per step of the range, a step without a call one with no calls. The app's
	// identity is put into the query here, as for QueryAppLogs.
	FunctionMetrics(ctx context.Context, db database.IDB, app *entity.App,
		q *FunctionMetricsQuery) (*logging.InvocationStatsResp, error)

	// ProxyHistory says whether the proxy's stored lines can be read, as
	// AppHistory does for an app's: an app's HTTP numbers are counted from them,
	// whatever the app's own log driver.
	ProxyHistory(ctx context.Context, db database.IDB) (*AppHistory, error)

	// HTTPMetrics counts an app's requests in the proxy's access log: one point
	// per step of the range, a step without a request one with none. The proxy's
	// identity and the app's services are put into the query here.
	HTTPMetrics(ctx context.Context, db database.IDB, app *entity.App,
		q *FunctionMetricsQuery) (*logging.HTTPStatsResp, error)

	// ResourceMetrics reads an app's containers' usage from the rows the agent
	// writes: one point per step, a step without a row one with none. The
	// agent's identity and the app's id are put into the query here.
	ResourceMetrics(ctx context.Context, db database.IDB, app *entity.App,
		q *FunctionMetricsQuery) (*logging.ResourceStatsResp, error)
}
