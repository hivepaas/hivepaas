package taskappdeploy

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appDeploymentService appdeploymentservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appDeploymentService appdeploymentservice.Service,
) *Executor {
	e := &Executor{
		appDeploymentService: appDeploymentService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeAppDeploy, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	execData *queue.TaskExecData,
) error {
	_, err := e.appDeploymentService.Deploy(ctx, db, &appdeploymentservice.AppDeploymentReq{
		TaskExecData: execData,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
