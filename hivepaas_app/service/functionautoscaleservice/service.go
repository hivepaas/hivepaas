// Package functionautoscaleservice scales functions' replicas from how busy
// their invocation lines say they are. See
// docs/superpowers/specs/2026-10-03-function-autoscale-design.md.
package functionautoscaleservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Service interface {
	// EnsureJob turns the periodic job that scales functions on while a
	// function has autoscale on, and off when none has: no function with
	// autoscale, no runs at all.
	EnsureJob(ctx context.Context, db database.IDB) error

	// Run is one run of that job: every function with autoscale on, one query,
	// a decision each. A run that scales one asks for its task to be saved.
	Run(ctx context.Context, data *queue.PeriodicExecData) error

	// Events are a function's scalings since a time, the latest first, at most
	// limit of them (0 for no limit): what the job's saved tasks say of it.
	Events(ctx context.Context, db database.IDB, appID string, since time.Time, limit int) ([]*Event, error)
}

// Event is one scaling of a function: when, from and to how many replicas,
// and what it was decided from.
type Event struct {
	Time      time.Time
	From      int
	To        int
	InFlight  float64
	Calls     int64
	Throttled int64
	Reason    string
}
