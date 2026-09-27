package schedjobtriggerservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

// Service runs the scheduled jobs that listen to what happens to an app.
type Service interface {
	// Fire schedules a run of every job listening to event of app. Its runs are
	// inserted in a transaction of its own: a caller in the middle of one may
	// wait for them.
	Fire(ctx context.Context, event base.SchedJobTriggerEvent, app *entity.App,
		info *TriggerInfo) (*FireResult, error)
	// WaitForRuns waits for runs to end, logging to logStore how each did. A run
	// that fails, is canceled or passes its timeout fails the wait; ctx ending
	// cancels the runs still going.
	WaitForRuns(ctx context.Context, runs []*Run, logStore *tasklog.Store) error
	// FireAppStatusEvents runs the jobs listening to the apps' status changes:
	// app-enabled or app-disabled, as each now is. Called once the change is
	// committed; failures are logged.
	FireAppStatusEvents(ctx context.Context, apps []*entity.App)
}
