package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UpdateNodeReq struct {
	settings.UpdateSettingReq
	Name         string                  `json:"name"`
	Labels       map[string]string       `json:"labels"`
	Availability docker.NodeAvailability `json:"availability"`
}

func NewUpdateNodeReq() *UpdateNodeReq {
	return &UpdateNodeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNodeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 1, nodeNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Availability, false, docker.AllNodeAvailabilities,
		"availability")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
