package imageuc

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/image"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc/imagedto"
)

func (uc *UC) ListImage(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.ListImageReq,
) (*imagedto.ListImageResp, error) {
	listResp, err := uc.dockerManager.ImageList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	filterImages := listResp.Items
	if req.Search != "" {
		keyword := strings.ToLower(req.Search)
		filterImages = gofn.FilterPtr(filterImages, func(img *image.Summary) bool {
			return gofn.Contain(img.RepoTags, keyword)
		})
	}
	if len(auth.AllowObjectIDs) > 0 {
		filterImages = gofn.FilterPtr(filterImages, func(img *image.Summary) bool {
			return gofn.Contain(auth.AllowObjectIDs, img.ID)
		})
	}

	return &imagedto.ListImageResp{
		Meta: &basedto.ListMeta{Page: &basedto.PagingMeta{
			Offset: 0,
			Limit:  req.Paging.Limit,
			Total:  len(filterImages),
		}},
		Data: imagedto.TransformImages(filterImages, false),
	}, nil
}
