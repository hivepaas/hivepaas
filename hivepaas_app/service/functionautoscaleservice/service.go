// Package functionautoscaleservice scales functions' replicas from how busy
// their invocation lines say they are. See
// docs/superpowers/specs/2026-10-03-function-autoscale-design.md.
package functionautoscaleservice

import (
	"context"

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
}
