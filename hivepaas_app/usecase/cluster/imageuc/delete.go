package imageuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc/imagedto"
)

func (uc *UC) DeleteImage(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.DeleteImageReq,
) (*imagedto.DeleteImageResp, error) {
	_, err := uc.dockerManager.ImageRemove(ctx, req.ImageID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &imagedto.DeleteImageResp{}, nil
}
