package commandservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

type Service interface {
	BuildCommandEnvVars(ctx context.Context, db database.IDB, app *entity.App, cmd *entity.CommandTemplate) (
		[]*envvarservice.EnvVar, error)
	GetCommand(ctx context.Context, cmdType, cmdName string) (*entity.Setting, error)
}
