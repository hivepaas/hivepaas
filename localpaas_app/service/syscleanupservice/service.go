package syscleanupservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	Cleanup(ctx context.Context, db database.Tx, req *SysCleanupReq) (*SysCleanupResp, error)
}
