package syscleanupservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	Cleanup(ctx context.Context, db database.Tx, req *SysCleanupReq) (*SysCleanupResp, error)
}
