package vulnignoredto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteVulnIgnoreReq struct {
	settings.DeleteSettingReq
}

func NewDeleteVulnIgnoreReq() *DeleteVulnIgnoreReq {
	return &DeleteVulnIgnoreReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteVulnIgnoreReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteVulnIgnoreResp struct {
	Meta *basedto.Meta `json:"meta"`
}
