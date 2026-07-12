package nodeuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
			nodeID := dockerhelper.ParseID(data.Setting.ID)
			inspect, err := uc.dockerManager.NodeInspect(ctx, nodeID)
			if err != nil {
				return apperrors.Wrap(err)
			}
			node := &inspect.Node
			spec := &node.Spec

			if req.Name != "" {
				spec.Annotations.Name = req.Name //nolint
			}
			spec.Labels = dockerhelper.ApplyUserLabels(spec.Labels, req.Labels)
			if req.Role != "" {
				spec.Role = swarm.NodeRole(req.Role)
			}
			if req.Availability != "" {
				spec.Availability = swarm.NodeAvailability(req.Availability)
			}

			_, err = uc.dockerManager.NodeUpdate(ctx, nodeID, &node.Version, spec)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodedto.UpdateNodeResp{}, nil
}
