package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateSchedJobReq struct {
	settings.UpdateSettingReq
	*SchedJobBaseReq
}

func NewUpdateSchedJobReq() *UpdateSchedJobReq {
	return &UpdateSchedJobReq{}
}

func (req *UpdateSchedJobReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSchedJobResp struct {
	Meta *basedto.Meta `json:"meta"`
}
