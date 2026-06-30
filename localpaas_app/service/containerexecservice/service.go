package containerexecservice

import (
	"context"
)

type Service interface {
	ContainerExec(ctx context.Context, req *ContainerExecReq) (*ContainerExecResp, error)
}
