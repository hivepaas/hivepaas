package approutingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ApplyRoutingSettings(ctx context.Context, db database.IDB, req *ApplyAppRoutingReq) (*ApplyAppRoutingResp, error)
}
