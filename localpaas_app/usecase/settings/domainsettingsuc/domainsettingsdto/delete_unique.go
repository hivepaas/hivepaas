package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type DeleteUniqueDomainSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteUniqueDomainSettingsReq() *DeleteUniqueDomainSettingsReq {
	return &DeleteUniqueDomainSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteUniqueDomainSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteUniqueDomainSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
