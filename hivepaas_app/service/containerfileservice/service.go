package containerfileservice

import (
	"context"
)

type Service interface {
	StreamFile(ctx context.Context, req *StreamFileReq) (*StreamFileResp, error)
}
