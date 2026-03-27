package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type DeleteFileReq struct {
	settings.DeleteSettingReq
}

func NewDeleteFileReq() *DeleteFileReq {
	return &DeleteFileReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteFileResp struct {
	Meta *basedto.Meta `json:"meta"`
}
