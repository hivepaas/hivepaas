package networkuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) GetNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.GetNetworkReq,
) (*networkdto.GetNetworkResp, error) {
	network, err := uc.dockerManager.NetworkInspect(ctx, req.NetworkID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if req.ProjectID != "" {
		project, err := uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		if network.Labels[docker.StackLabelNamespace] != project.Key {
			return nil, apperrors.NewNotFound("Network").WithMsgLog("network not belong to project")
		}
	}

	return &networkdto.GetNetworkResp{
		Data: networkdto.TransformNetworkInspection(network),
	}, nil
}
