package loggingservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/services/logging"
)

// Status is what the logging stack is currently doing.
type Status struct {
	Enabled          bool
	BackendServiceID string
	CollectorService string
	BackendReady     bool
	// ExcludedApps are apps whose logs will not reach the backend usable.
	ExcludedApps []ExcludedApp
}

// ExcludedReason is why an app is not collected.
type ExcludedReason string

const (
	// ExcludedReasonDriverUnreadable is a log driver the collector cannot read:
	// `local`, which new apps used until json-file became the default, or one
	// the operator chose, such as syslog.
	ExcludedReasonDriverUnreadable ExcludedReason = "driver-unreadable"

	// ExcludedReasonIdentityMissing is a readable driver whose lines do not carry
	// this app's identity - an app created before the identity labels existed,
	// or cloned before clones were copied deeply. Its lines are collected but
	// cannot be scoped to the app, which for the read path is as good as absent.
	ExcludedReasonIdentityMissing ExcludedReason = "identity-missing"
)

// ExcludedApp is one app that logging will not show.
type ExcludedApp struct {
	AppID  string
	Name   string
	Reason ExcludedReason
	// Driver is the log driver the service runs with, empty for the daemon's.
	Driver string
}

// AppLogQuery is a search over one app's stored logs. The app is not a field:
// the service puts it into the query itself, which is the point.
type AppLogQuery struct {
	Contains string
	Levels   []string
	Streams  []string
	Start    time.Time
	End      time.Time
	Limit    int
}

// HistoryUnavailableReason is why an app's stored logs cannot be shown.
type HistoryUnavailableReason string

const (
	HistoryReasonDisabled         HistoryUnavailableReason = "disabled"
	HistoryReasonAppsNotCollected HistoryUnavailableReason = "apps-not-collected"
	HistoryReasonNoQueryEndpoint  HistoryUnavailableReason = "no-query-endpoint"
	HistoryReasonDriverUnreadable HistoryUnavailableReason = "driver-unreadable"
	HistoryReasonIdentityMissing  HistoryUnavailableReason = "identity-missing"
)

// AppHistory says whether an app's stored logs can be shown.
type AppHistory struct {
	Available bool
	Reason    HistoryUnavailableReason
}

type Service interface {
	// Apply makes the cluster match the stored configuration, deploying or
	// removing as needed.
	Apply(ctx context.Context, db database.IDB) error

	// TearDown removes the collector and the backend, keeping the data volume.
	TearDown(ctx context.Context) error

	Status(ctx context.Context, db database.IDB) (*Status, error)

	// QueryAppLogs searches one app's stored logs. The app's identity is put
	// into the query here; nothing in q can widen it.
	QueryAppLogs(ctx context.Context, db database.IDB, app *entity.App, q *AppLogQuery) (*logging.QueryResp, error)

	// AppHistory says whether stored logs can be shown for the app, and if not,
	// why - so the dashboard can say so instead of showing an empty list.
	AppHistory(ctx context.Context, db database.IDB, app *entity.App) (*AppHistory, error)
}
