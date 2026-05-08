package appdeploymentservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	Deploy(ctx context.Context, db database.Tx, data *DeploymentData) (err error)
}
