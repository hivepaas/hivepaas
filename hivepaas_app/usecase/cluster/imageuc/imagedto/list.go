package imagedto

import (
	"github.com/moby/moby/api/types/image"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListImageReq struct {
	Search string `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListImageReq() *ListImageReq {
	return &ListImageReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
		},
	}
}

func (req *ListImageReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListImageResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*ImageResp      `json:"data"`
}

func TransformImages(images []image.Summary, detailed bool) []*ImageResp {
	return gofn.MapSlice(images, func(img image.Summary) *ImageResp {
		return TransformImage(&img, detailed)
	})
}
