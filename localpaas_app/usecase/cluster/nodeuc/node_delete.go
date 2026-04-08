package nodeuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) DeleteNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.DeleteNodeReq,
) (*nodedto.DeleteNodeResp, error) {
	var options []docker.NodeRemoveOption
	if req.Force {
		options = append(options, docker.NodeRemoveForce(true))
	}

	err := uc.dockerManager.NodeRemove(ctx, req.NodeID, options...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodedto.DeleteNodeResp{}, nil
}
