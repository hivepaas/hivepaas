package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type DeleteDomainSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteDomainSettingsReq() *DeleteDomainSettingsReq {
	return &DeleteDomainSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteDomainSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteDomainSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
