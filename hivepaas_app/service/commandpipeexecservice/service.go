package commandpipeexecservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	CommandPipeExec(ctx context.Context, db database.IDB, req *CommandPipeExecReq) (*CommandPipeExecResp, error)
}
