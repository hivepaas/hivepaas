package imageuc

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc/imagedto"
)

func (uc *UC) ListImage(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.ListImageReq,
) (*imagedto.ListImageResp, error) {
	images, err := uc.dockerManager.ImageList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	filterImages := images
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
