package loggingservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type SettingApplyReq struct {
	Setting *entity.Setting // if nil, it will be loaded from DB
}

type SettingApplyResp struct {
}

// Status is what the logging stack is currently doing.
type Status struct {
	Enabled            bool
	CollectorServiceID string
	CollectorReady     bool
	BackendServiceID   string
	BackendReady       bool
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
