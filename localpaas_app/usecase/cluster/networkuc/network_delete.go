package networkuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) DeleteNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.DeleteNetworkReq,
) (*networkdto.DeleteNetworkResp, error) {
	if req.ProjectID != "" {
		project, err := uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		network, err := uc.dockerManager.NetworkInspect(ctx, req.NetworkID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		if network.Labels[docker.StackLabelNamespace] != project.Key {
			return nil, apperrors.NewNotFound("Network").WithMsgLog("network not belong to project")
		}
	}

	err := uc.dockerManager.NetworkRemove(ctx, req.NetworkID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &networkdto.DeleteNetworkResp{}, nil
}
