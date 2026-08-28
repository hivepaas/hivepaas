package nodeuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func (uc *UC) UpdateNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.UpdateNodeReq,
) (*nodedto.UpdateNodeResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			nodeID := data.Setting.MustAsClusterNode().RefID
			inspect, err := uc.dockerManager.NodeInspect(ctx, nodeID)
			if err != nil {
				return hperrors.Wrap(err)
			}
			node := &inspect.Node
			spec := &node.Spec

			if req.Name != "" {
				spec.Name = req.Name
			}
			spec.Labels = dockerhelper.ApplyUserLabels(spec.Labels, req.Labels)
			if req.Availability != "" {
				spec.Availability = swarm.NodeAvailability(req.Availability)
			}

			_, err = uc.dockerManager.NodeUpdate(ctx, nodeID, &node.Version, spec)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &nodedto.UpdateNodeResp{}, nil
}
