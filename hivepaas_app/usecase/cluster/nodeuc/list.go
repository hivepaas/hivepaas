package nodeuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) ListNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.ListNodeReq,
) (_ *nodedto.ListNodeResp, err error) {
	var currNodes []swarm.Node
	if req.Scope.IsGlobalScope() {
		currNodes, err = uc.clusterService.SyncNodes(ctx, uc.DB)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listNodesInDocker(ctx, resp.Data, currNodes, refClusterObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := nodedto.TransformNodes(resp.Data, resp.RefObjects, refClusterObjects, false)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &nodedto.ListNodeResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}

func (uc *UC) listNodesInDocker(
	ctx context.Context,
	settings []*entity.Setting,
	currNodes []swarm.Node,
	refClusterObjects *entity.RefClusterObjects,
) error {
	if currNodes == nil {
		nodes := make([]string, 0, len(settings))
		for _, setting := range settings {
			nodes = append(nodes, dockerhelper.ParseID(setting.ID))
		}
		if len(nodes) == 0 {
			return nil
		}

		res, err := uc.dockerManager.NodeListByIDs(ctx, nodes)
		if err != nil {
			return apperrors.New(err)
		}
		currNodes = res.Items
	}

	for i := range currNodes {
		node := &currNodes[i]
		refClusterObjects.RefNodes[node.ID] = node
	}
	return nil
}
