package nodeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/services/docker"
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

	_, err := uc.dockerManager.NodeRemove(ctx, req.NodeID, options...)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &nodedto.DeleteNodeResp{}, nil
}
