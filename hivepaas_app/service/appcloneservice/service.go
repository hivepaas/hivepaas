package appcloneservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	CreateAppCloneTask(app *entity.App) (*entity.Task, error)

	CloneApp(ctx context.Context, db database.IDB, req *AppCloneReq) (*AppCloneResp, error)
}
