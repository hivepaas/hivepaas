package healthcheckserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

// healthEvent is the event of a check's result: health-down or health-up when
// its state changed from one it knew; none on its first run or when its last
// state was lost, which the task is saved for but is no change of health.
func healthEvent(
	last *cacheentity.HealthcheckState,
	status base.TaskStatus,
) (base.SchedJobTriggerEvent, bool) {
	if last == nil {
		return "", false
	}
	wasUp := last.State == base.HealthcheckStateSuccess
	isUp := status == base.TaskStatusDone
	switch {
	case wasUp && !isUp:
		return base.SchedJobTriggerHealthDown, true
	case !wasUp && isUp:
		return base.SchedJobTriggerHealthUp, true
	}
	return "", false
}

// fireHealthEvent runs the jobs listening to the change of the app's health. A
// failure is logged: the check itself went as it went.
func (s *service) fireHealthEvent(ctx context.Context, data *healthcheckData) {
	event, changed := healthEvent(data.LastHealthcheckState, data.Task.Status)
	if !changed || data.Scope == nil {
		return
	}
	app := data.Scope.GetApp()
	if app == nil {
		return
	}
	if _, err := s.schedJobTriggerService.Fire(context.WithoutCancel(ctx), event, app, nil); err != nil {
		logging.GlobalLogger().Errorf("failed to start the jobs listening to %s of app %s: %v", event, app.ID, err)
	}
}
