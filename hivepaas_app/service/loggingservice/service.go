package loggingservice

import (
	"context"
	"time"

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

	// FunctionLoad says how busy functions were over [start, end), by app id,
	// in one query: what autoscale decides from; and over its last part from
	// shortStart, apart - zero for none. It fails, rather than answer nothing,
	// when the logs cannot be read: no data is not no load.
	FunctionLoad(ctx context.Context, db database.IDB, appIDs []string,
		start, end, shortStart time.Time) (map[string]*logging.InvocationLoad, error)

	// RequestLoad says how busy apps were over [start, end) by the proxy's
	// access log, by app id, in one query; and over its last part from
	// shortStart, apart. It fails when the logs cannot be read, as
	// FunctionLoad does.
	RequestLoad(ctx context.Context, db database.IDB, appIDs []string,
		start, end, shortStart time.Time) (map[string]*logging.RequestLoad, error)

	// CPULoad reads apps' containers' CPU over [start, end) from the agent's
	// rows, by app id, in one query. It fails when the logs cannot be read.
	CPULoad(ctx context.Context, db database.IDB, appIDs []string,
		start, end time.Time) (map[string][]*logging.ContainerCPU, error)

	// RouteMetrics sums an app's routes from the rows the agent writes from
	// OBI on each node: by step, a step without a row left out; and by kind,
	// method and route, the busiest first. The agent's identity and the app's
	// id are put into the query here.
	RouteMetrics(ctx context.Context, db database.IDB, app *entity.App,
		q *FunctionMetricsQuery) (*logging.OBIStatsResp, error)

	// DependencyMetrics sums an app's calls as RouteMetrics its routes: each
	// step split by the calls' kind; and by kind, peer, method and operation.
	DependencyMetrics(ctx context.Context, db database.IDB, app *entity.App,
		q *FunctionMetricsQuery) (*logging.OBIStatsResp, error)

	// PerformanceStatus reads each node's latest status about OBI, by node id,
	// from the rows its agent writes while the logs are stored - every minute
	// while the feature is on, every 10 while it is off - over the last since:
	// a node with none has no agent writing them.
	PerformanceStatus(ctx context.Context, db database.IDB, since time.Duration) (
		map[string]*PerformanceNodeStatus, error)

	// Sync brings the logging apps to the stored settings, as a save does, and
	// further: an app the settings no longer want is removed, its stored logs
	// kept; one whose service is gone is removed and provisioned again. It then
	// checks OBI on the nodes, and takes the nodes no longer in the cluster off
	// the ones that run it. It runs inside the caller's transaction, and locks
	// the settings as a save does.
	Sync(ctx context.Context, db database.IDB) (*SyncResp, error)
}
