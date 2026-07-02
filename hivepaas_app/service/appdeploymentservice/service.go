package appdeploymentservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	CreateDeploymentAndTask(app *entity.App, deploymentSettings *entity.AppDeploymentSettings) (
		*entity.Deployment, *entity.Task, error)

	Deploy(ctx context.Context, db database.Tx, req *AppDeploymentReq) (*AppDeploymentResp, error)
}
