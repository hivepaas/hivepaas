package settingeventserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

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
	tasks = gofn.ToSliceSkippingNil(tasks...)
	if len(tasks) == 0 {
		return
	}
	// A task the queue is not told of now is found by its own scan.
	_ = s.taskQueue.ScheduleTask(ctx, tasks...)
}
