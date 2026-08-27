package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteDomainSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteDomainSettingsReq() *DeleteDomainSettingsReq {
	return &DeleteDomainSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteDomainSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteDomainSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
