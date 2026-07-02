package schedjobexecservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	SchedJobExec(ctx context.Context, db database.Tx, req *SchedJobExecReq) (*SchedJobExecResp, error)
}
