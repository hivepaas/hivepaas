package taskappclone

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appCloneService appcloneservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appCloneService appcloneservice.Service,
) *Executor {
	e := &Executor{
		appCloneService: appCloneService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeAppClone, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	execData *queue.TaskExecData,
) error {
	_, err := e.appCloneService.CloneApp(ctx, db, &appcloneservice.AppCloneReq{
		TaskExecData: execData,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
