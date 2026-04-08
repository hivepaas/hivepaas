package volumeuc

import (
	"context"
	"encoding/json"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) GetVolumeInspection(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.GetVolumeInspectionReq,
) (*volumedto.GetVolumeInspectionResp, error) {
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

	resp, err := json.MarshalIndent(volume, "", "   ")
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &volumedto.GetVolumeInspectionResp{
		Data: reflectutil.UnsafeBytesToStr(resp),
	}, nil
}
