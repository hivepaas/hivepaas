package volumeservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc/volumeagentdto"
)

func RemoveVolume(
	ctx context.Context,
	uc *volumeagentuc.UC,
	req *agentproto.RemoveVolumeReq,
) (*agentproto.RemoveVolumeResp, error) {
	if req == nil {
		return &agentproto.RemoveVolumeResp{}, nil
	}

	_, err := uc.RemoveVolume(ctx, &volumeagentdto.RemoveVolumeReq{
		VolumeID: req.GetVolumeId(),
		Force:    req.GetForce(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &agentproto.RemoveVolumeResp{}, nil
}
