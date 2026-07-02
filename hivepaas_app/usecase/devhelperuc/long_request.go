package devhelperuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) LongRequest(
	ctx context.Context,
	auth *basedto.Auth,
	req *devhelperdto.LongRequestReq,
) (*devhelperdto.LongRequestResp, error) {
	time.Sleep(req.Duration.ToDuration())
	return &devhelperdto.LongRequestResp{}, nil
}
