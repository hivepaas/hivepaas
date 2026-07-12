package imageuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc/imagedto"
)

func (uc *UC) GetImage(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.GetImageReq,
) (*imagedto.GetImageResp, error) {
	inspect, err := uc.dockerManager.ImageInspect(ctx, req.ImageID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagedto.GetImageResp{
		Data: imagedto.TransformImageFromResp(&inspect.InspectResponse, true),
	}, nil
}
