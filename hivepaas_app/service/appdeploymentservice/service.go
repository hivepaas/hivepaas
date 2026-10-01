package appdeploymentservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	CreateDeploymentAndTask(app *entity.App, deploymentSettings *entity.AppDeploymentSettings,
		args DeploymentArgs) (*entity.Deployment, *entity.Task, error)

	Deploy(ctx context.Context, db database.Tx, req *AppDeploymentReq) (*AppDeploymentResp, error)

	// CheckBuildSource checks what a build will need before its settings are
	// saved: a registry for its image when the cluster has several nodes, and
	// the repository and ref it checks out, if it checks out one.
	CheckBuildSource(ctx context.Context, req *CheckBuildSourceReq) error
}
