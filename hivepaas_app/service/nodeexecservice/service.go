package nodeexecservice

import (
	"context"
)

type Service interface {
	ExecCommand(ctx context.Context, req *CommandExecReq) (*CommandExecResp, error)
}
