package nodedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UpdateNodeReq struct {
	settings.UpdateSettingReq
	Name         string                  `json:"name"`
	Labels       map[string]string       `json:"labels"`
	Role         docker.NodeRole         `json:"role"`
	Availability docker.NodeAvailability `json:"availability"`
	UpdateVer    int                     `json:"updateVer"`
}

func NewUpdateNodeReq() *UpdateNodeReq {
	return &UpdateNodeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 1, nodeNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Role, false, docker.AllNodeRoles, "role")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Availability, false, docker.AllNodeAvailabilities,
		"availability")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
