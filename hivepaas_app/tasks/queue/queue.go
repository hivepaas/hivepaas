package queue

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type TaskQueue interface {
	Start() error
	Shutdown() error

	StartScheduler() error     // resume from pause
	StartAllSchedulers() error // resume from pause
	StopScheduler() error      // pause the scheduler
	StopAllSchedulers() error  // pause the scheduler

	RegisterExecutor(typ base.TaskType, execFunc TaskExecFunc)
	RegisterHealthcheckExecutor(execFunc HealthcheckExecFunc)

	ScheduleTask(ctx context.Context, tasks ...*entity.Task) error
	UnscheduleTask(ctx context.Context, tasks ...*entity.Task) error
	ScheduleTasksForSchedJobs(ctx context.Context, db database.Tx, schedJobs []*entity.Setting,
		unscheduleCurrentTasks bool) error
	ScheduleTasksForSchedJob(ctx context.Context, db database.Tx, schedJob *entity.Setting,
		unscheduleCurrentTasks bool) error
}
