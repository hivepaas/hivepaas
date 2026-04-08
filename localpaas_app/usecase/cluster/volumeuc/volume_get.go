package volumeuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) GetVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.GetVolumeReq,
) (*volumedto.GetVolumeResp, error) {
	volume, _, err := uc.dockerManager.VolumeInspect(ctx, req.VolumeID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if req.ProjectID != "" {
		project, err := uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		if volume.Labels[docker.StackLabelNamespace] != project.Key {
			return nil, apperrors.NewNotFound("Volume").WithMsgLog("volume not belong to project")
		}
	}

	return &volumedto.GetVolumeResp{
		Data: volumedto.TransformVolume(volume, true),
	}, nil
}
