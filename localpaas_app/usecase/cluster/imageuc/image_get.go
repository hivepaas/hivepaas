package imageuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc/imagedto"
)

func (uc *UC) GetImage(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.GetImageReq,
) (*imagedto.GetImageResp, error) {
	img, err := uc.dockerManager.ImageInspect(ctx, req.ImageID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagedto.GetImageResp{
		Data: imagedto.TransformImageFromResp(img, true),
	}, nil
}
