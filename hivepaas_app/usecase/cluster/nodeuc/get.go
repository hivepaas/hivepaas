package nodeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) GetNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.GetNodeReq,
) (*nodedto.GetNodeResp, error) {
	resp, err := uc.dockerManager.NodeInspect(ctx, req.NodeID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &nodedto.GetNodeResp{
		Data: nodedto.TransformNode(&resp.Node, true),
	}, nil
}
