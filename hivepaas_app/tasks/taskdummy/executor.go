package taskdummy

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
}

func NewExecutor(
	taskQueue queue.TaskQueue,
) *Executor {
	e := &Executor{}
	taskQueue.RegisterExecutor(base.TaskTypeDummy, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	_ database.Tx,
	task *queue.TaskExecData,
) (err error) {
	args := gofn.Must(task.Task.ArgsAsDummy())
	time.Sleep(args.Sleep.ToDuration())
	return nil
}
