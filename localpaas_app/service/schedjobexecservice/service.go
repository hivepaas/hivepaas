package schedjobexecservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	SchedJobExec(ctx context.Context, db database.Tx, req *SchedJobExecReq) (*SchedJobExecResp, error)
}
