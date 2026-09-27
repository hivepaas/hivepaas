package schedjobtriggerservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// Service runs the scheduled jobs that listen to what happens to an app.
type Service interface {
	// Fire schedules a run of every job listening to event of app. Its runs are
	// inserted in a transaction of its own: a caller in the middle of one may
	// wait for them.
	Fire(ctx context.Context, event base.SchedJobTriggerEvent, app *entity.App,
		info *TriggerInfo) (*FireResult, error)
}
