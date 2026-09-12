package loggingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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

type Service interface {
	// Apply makes the cluster match the stored configuration, deploying or
	// removing as needed.
	Apply(ctx context.Context, db database.IDB) error

	// TearDown removes the collector and the backend, keeping the data volume.
	TearDown(ctx context.Context) error

	Status(ctx context.Context, db database.IDB) (*Status, error)
}
