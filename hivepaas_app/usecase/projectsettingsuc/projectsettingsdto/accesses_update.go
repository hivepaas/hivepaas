package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateUserAccessesReq struct {
	ProjectID       string                `json:"-"`
	UserAccesses    []*UserAccessReq      `json:"userAccesses"`
	EnvUserAccesses []*EnvUserAccessesReq `json:"envUserAccesses"`
	UpdateVer       int                   `json:"updateVer"`
}

type UserAccessReq struct {
	ID     string             `json:"id"`
	Access base.AccessActions `json:"access"`
}

type EnvUserAccessesReq struct {
	Name         string           `json:"name"`
	UserAccesses []*UserAccessReq `json:"userAccesses"`
}

func NewUpdateUserAccessesReq() *UpdateUserAccessesReq {
	return &UpdateUserAccessesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateUserAccessesReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateUserAccessesResp struct {
	Meta *basedto.Meta `json:"meta"`
}
