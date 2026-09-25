package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// recordMountRefresh records, in the event's transaction, a refresh of the apps
// that mount the setting - or of the entry's own app - into tasks.
func (s *service) recordMountRefresh(
	ctx context.Context, db database.IDB, tasks *[]*entity.Task, setting *entity.Setting,
) error {
	task, err := s.settingMountService.RecordRefresh(ctx, db, setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if task != nil {
		*tasks = append(*tasks, task)
	}
	return nil
}

func (s *service) ScheduleTasks(ctx context.Context, tasks ...*entity.Task) {
	s.settingMountService.Schedule(ctx, tasks...)
}
