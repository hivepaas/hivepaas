package imagedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DeleteImageReq struct {
	ImageID string `json:"-"`
	Force   bool   `json:"-" mapstructure:"force"`
}

func NewDeleteImageReq() *DeleteImageReq {
	return &DeleteImageReq{}
}

func (req *DeleteImageReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: image id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.ImageID, true, 1, imageIDMaxLen, "imageID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteImageResp struct {
	Meta *basedto.Meta `json:"meta"`
}
