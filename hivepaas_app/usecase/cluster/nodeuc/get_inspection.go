package nodeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) GetNodeInspection(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.GetNodeInspectionReq,
) (*nodedto.GetNodeInspectionResp, error) {
	resp, err := uc.dockerManager.NodeInspect(ctx, req.NodeID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &nodedto.GetNodeInspectionResp{
		Data: reflectutil.UnsafeBytesToStr(resp.Raw),
	}, nil
}
