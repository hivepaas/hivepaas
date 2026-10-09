package schedjobuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// schedulingQueue keeps the jobs it was asked to schedule.
type schedulingQueue struct {
	queue.TaskQueue
	scheduled []*entity.Setting
}

func (q *schedulingQueue) ScheduleTasksForSchedJob(
	_ context.Context, _ database.Tx, job *entity.Setting, _ bool,
) error {
	q.scheduled = append(q.scheduled, job)
	return nil
}

// A job turned on is scheduled as it is now - active - so that its runs are
// made at once, not at the next scan, minutes later.
func TestAJobTurnedOnIsScheduledAsItIsNow(t *testing.T) {
	tasks := &schedulingQueue{}
	uc := &UC{taskQueue: tasks}
	before := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob, Status: base.SettingStatusDisabled}
	after := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob, Status: base.SettingStatusActive}

	err := uc.scheduleWithStatus(context.Background(), database.Tx{},
		&settings.UpdateSettingStatusData{Setting: before}, &settings.PersistingSettingStatusData{Setting: after})

	assert.NoError(t, err)
	assert.Equal(t, []*entity.Setting{after}, tasks.scheduled)
}
