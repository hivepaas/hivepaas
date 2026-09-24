package loggingservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/services/logging"
)

type SettingApplyReq struct {
	Setting *entity.Setting // if nil, it will be loaded from DB

	// BackendResources is what the managed backend runs under. It is not stored
	// with the settings: the backend's service is where it lives, the same place
	// the app's own resource screen reads and writes. Nil leaves a running
	// backend as it is, and gives a new one the defaults.
	BackendResources *logging.Resources

	TriggerUserID string

	// RemoveApp and RemoveStorage are asked of this request, not stored by it.
	// Switching logging off - or handing the backend or the collector to a system
	// HivePaaS does not run - takes an app down, and the confirmation arrives here.
	RemoveApp bool
	// RemoveStorage deletes the stored logs with the backend: its own directory
	// inside the volume, not the volume.
	RemoveStorage bool
}

type SettingApplyResp struct {
	// Tasks are the deployments this call queued. They are left unscheduled: a
	// task row can be picked up only once the transaction it was written in has
	// committed, and Apply runs inside the caller's.
	Tasks []*entity.Task

	// Cleanup removes from docker what provisioning created there, for the caller
	// to run when its transaction does not commit. It is set when this call
	// provisioned an app, even if provisioning failed.
	Cleanup func(ctx context.Context) error

	// RemovedApps says this call took at least one app down.
	RemovedApps bool
}

// Status is what the logging stack is currently doing.
type Status struct {
	Enabled bool

	// Backend and Collector are the apps HivePaaS runs, nil when it runs none.
	Backend   *AppStatus
	Collector *AppStatus

	// BackendResources is what the backend's service runs under, or what a new
	// backend would get when there is none.
	BackendResources logging.Resources
}

// AppStatus is one app of the stack, as the settings screen shows it.
type AppStatus struct {
	AppID     string
	ProjectID string
	// ProjectEnv is the key of the app's environment, which its screen's address
	// carries.
	ProjectEnv string
	// RunningTasks and DesiredTasks count containers: one for the backend, one
	// per node for the collector.
	RunningTasks uint64
	DesiredTasks uint64
}

// Ready says the app runs every container it should.
func (s *AppStatus) Ready() bool {
	return s != nil && s.DesiredTasks > 0 && s.RunningTasks >= s.DesiredTasks
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

// AppLogQuery is a search over one app's stored logs. The app is not a field:
// the service puts it into the query itself, which is the point.
type AppLogQuery struct {
	Search  *logging.TextSearch
	Levels  []string
	Streams []string
	Start   time.Time
	End     time.Time
	Limit   int
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
	// Retention is how far back the stored logs reach. It is zero when the
	// backend is not one HivePaaS keeps: see retentionOf.
	Retention timeutil.Duration
}
