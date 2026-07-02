package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteNetworkReq struct {
	settings.DeleteSettingReq
}

func NewDeleteNetworkReq() *DeleteNetworkReq {
	return &DeleteNetworkReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
}
