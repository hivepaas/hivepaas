package imagedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteImageReq struct {
	ImageID string `json:"-"`
	Force   bool   `json:"-" mapstructure:"force"`
}

func NewDeleteImageReq() *DeleteImageReq {
	return &DeleteImageReq{}
}

func (req *DeleteImageReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	// NOTE: image id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.ImageID, true, 1, imageIDMaxLen, "imageId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteImageResp struct {
	Meta *basedto.Meta `json:"meta"`
}
