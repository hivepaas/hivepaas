package volumeagentuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc/volumeagentdto"
)

// RemoveVolume removes a volume from the docker daemon this agent runs beside.
//
// A volume created by HivePaaS exists only on the node its data is on, and only
// once a task has actually mounted it there, so this is the one way to remove
// one that lives on a node other than the manager.
//
// A volume that is already gone is success, the same as it is for a removal
// issued locally: the caller asked for it not to be there.
func (uc *UC) RemoveVolume(
	ctx context.Context,
	req *volumeagentdto.RemoveVolumeReq,
) (*volumeagentdto.RemoveVolumeResp, error) {
	if req == nil || req.VolumeID == "" {
		return &volumeagentdto.RemoveVolumeResp{}, nil
	}

	uc.logger.Infof("RemoveVolume started: %s, force: %v", req.VolumeID, req.Force)

	_, err := uc.dockerManager.VolumeRemove(ctx, req.VolumeID, req.Force)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	return &volumeagentdto.RemoveVolumeResp{}, nil
}
