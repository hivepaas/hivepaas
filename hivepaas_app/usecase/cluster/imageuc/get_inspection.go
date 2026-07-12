package imageuc

import (
	"context"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc/imagedto"
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
