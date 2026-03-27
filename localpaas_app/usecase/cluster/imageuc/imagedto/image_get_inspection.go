package imagedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetImageInspectionReq struct {
	ImageID string `json:"-"`
}

func NewGetImageInspectionReq() *GetImageInspectionReq {
	return &GetImageInspectionReq{}
}

func (req *GetImageInspectionReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: node id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.ImageID, true, 1, imageIDMaxLen, "imageID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetImageInspectionResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data string        `json:"data"`
}
