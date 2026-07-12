package nodeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) GetNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.GetNodeReq,
) (*nodedto.GetNodeResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	refClusterObjects := entity.NewRefClusterObjects()
	err = uc.listNodesInDocker(ctx, []*entity.Setting{resp.Data}, nil, refClusterObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := nodedto.TransformNode(resp.Data, resp.RefObjects, refClusterObjects, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodedto.GetNodeResp{
		Data: respData,
	}, nil
}
