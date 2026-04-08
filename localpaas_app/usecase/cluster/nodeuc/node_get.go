package nodeuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) GetNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.GetNodeReq,
) (*nodedto.GetNodeResp, error) {
	node, _, err := uc.dockerManager.NodeInspect(ctx, req.NodeID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodedto.GetNodeResp{
		Data: nodedto.TransformNode(node, true),
	}, nil
}
