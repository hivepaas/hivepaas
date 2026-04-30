package devhelperuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) LongRequest(
	ctx context.Context,
	auth *basedto.Auth,
	req *devhelperdto.LongRequestReq,
) (*devhelperdto.LongRequestResp, error) {
	time.Sleep(req.Duration.ToDuration())
	return &devhelperdto.LongRequestResp{}, nil
}
