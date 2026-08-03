package appcloneservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	CloneApp(ctx context.Context, db database.Tx, req *AppCloneReq) (*AppCloneResp, error)
}
