package schedjobexecservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	SchedJobExec(ctx context.Context, db database.Tx, req *SchedJobExecReq) (*SchedJobExecResp, error)
	// RunCommand runs a command in an app outside any job, reading Stdin: a
	// restore's. Its output goes to the task's log.
	RunCommand(ctx context.Context, db database.Tx, req *RunCommandReq) (*RunCommandResp, error)
}
