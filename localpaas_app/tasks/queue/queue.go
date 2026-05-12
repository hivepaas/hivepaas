package queue

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
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
	ScheduleTasksForCronJob(ctx context.Context, db database.Tx, cronJob *entity.Setting,
		unscheduleCurrentTasks bool) error
}
