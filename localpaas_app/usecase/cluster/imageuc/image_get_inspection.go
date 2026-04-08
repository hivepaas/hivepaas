package imageuc

import (
	"context"
	"encoding/json"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc/imagedto"
)

func (uc *UC) GetImageInspection(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.GetImageInspectionReq,
) (*imagedto.GetImageInspectionResp, error) {
	img, err := uc.dockerManager.ImageInspect(ctx, req.ImageID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := json.MarshalIndent(img, "", "   ")
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagedto.GetImageInspectionResp{
		Data: reflectutil.UnsafeBytesToStr(resp),
	}, nil
}
