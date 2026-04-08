package networkuc

import (
	"context"
	"encoding/json"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) GetNetworkInspection(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.GetNetworkInspectionReq,
) (*networkdto.GetNetworkInspectionResp, error) {
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

	resp, err := json.MarshalIndent(network, "", "   ")
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &networkdto.GetNetworkInspectionResp{
		Data: reflectutil.UnsafeBytesToStr(resp),
	}, nil
}
