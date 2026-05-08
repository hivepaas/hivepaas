package taskappdeploy

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
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
	err := e.appDeploymentService.Deploy(ctx, db, &appdeploymentservice.DeploymentData{
		TaskExecData: execData,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
