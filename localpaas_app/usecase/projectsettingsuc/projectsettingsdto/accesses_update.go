package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateUserAccessesReq struct {
	ProjectID    string           `json:"-"`
	UserAccesses []*UserAccessReq `json:"userAccesses"`
}

type UserAccessReq struct {
	ID     string             `json:"id"`
	Access base.AccessActions `json:"access"`
}

func NewUpdateUserAccessesReq() *UpdateUserAccessesReq {
	return &UpdateUserAccessesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateUserAccessesReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	// TODO: validate user accesses
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateUserAccessesResp struct {
	Meta *basedto.Meta `json:"meta"`
}
