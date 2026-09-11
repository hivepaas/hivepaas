package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) UpdateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.UpdateVolumeReq,
) (*volumedto.UpdateVolumeResp, error) {
	req.Type = currentSettingType
	req.Auth = auth

	// Resolved out here so "current" fails before the transaction opens, and so
	// the pinning written below is a node id rather than a marker.
	pinning := req.Pinning()
	if pinning != nil {
		nodeID, err := uc.resolveCurrentNode(ctx, pinning.NodeID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		pinning.NodeID = nodeID
	}

	// NOTE: besides the node pinning, only `inheritable` and `default` are updatable
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			// Nothing said about the pinning leaves it exactly as it is - the
			// same rule the volume sync follows, and for the same reason: this
			// field is the operator's answer and nothing else may guess at it.
			if pinning == nil {
				return nil
			}

			current, err := data.Setting.AsClusterVolume()
			if err != nil {
				return hperrors.Wrap(err)
			}
			if current == nil {
				current = &entity.ClusterVolume{}
			}
			// Read, change, write back rather than replacing the payload, so a
			// field added to ClusterVolume later is not quietly dropped here.
			current.NodeID = pinning.NodeID
			current.NodeLabel = pinning.NodeLabel

			return hperrors.Wrap(pData.Setting.SetData(current))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.UpdateVolumeResp{}, nil
}
