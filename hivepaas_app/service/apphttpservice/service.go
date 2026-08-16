package apphttpservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ApplyHttpSettings(ctx context.Context, db database.IDB, req *ApplyAppHttpReq) (*ApplyAppHttpResp, error)
}
