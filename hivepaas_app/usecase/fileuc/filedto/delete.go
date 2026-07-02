package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteFileReq struct {
	ID string `json:"-" mapstructure:"-"`
}

func NewDeleteFileReq() *DeleteFileReq {
	return &DeleteFileReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteFileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
