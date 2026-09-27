package schedjobtriggerserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

func (s *service) FireAppStatusEvents(ctx context.Context, apps []*entity.App) {
	for _, app := range apps {
		event, ok := appStatusEvent(app.Status)
		if !ok {
			continue
		}
		if _, err := s.Fire(ctx, event, app, nil); err != nil {
			logging.GlobalLogger().Errorf("failed to start the jobs listening to %s of app %s: %v",
				event, app.ID, err)
		}
	}
}

// appStatusEvent is the event of an app's new status: app-enabled or
// app-disabled; none for a status that is no change of that kind.
func appStatusEvent(status base.AppStatus) (base.SchedJobTriggerEvent, bool) {
	switch status { //nolint:exhaustive // only these two are a change of status
	case base.AppStatusActive:
		return base.SchedJobTriggerAppEnabled, true
	case base.AppStatusDisabled:
		return base.SchedJobTriggerAppDisabled, true
	}
	return "", false
}
