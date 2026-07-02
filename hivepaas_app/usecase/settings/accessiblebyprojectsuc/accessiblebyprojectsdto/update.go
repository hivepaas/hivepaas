package accessiblebyprojectsdto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAccessibleByProjectsReq struct {
	SettingID            string                    `json:"-"`
	AccessibleByProjects []*AccessibleByProjectReq `json:"accessibleByProjects"`
}

type AccessibleByProjectReq struct {
	ID string `json:"id"`
}

func (req *AccessibleByProjectReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateID(&req.ID, true, field+"id")...)
	return res
}

func NewUpdateAccessibleByProjectsReq() *UpdateAccessibleByProjectsReq {
	return &UpdateAccessibleByProjectsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAccessibleByProjectsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	for i, r := range req.AccessibleByProjects {
		validators = append(validators, r.validate(fmt.Sprintf("accessibleByProjects[%d]", i))...)
	}
	validators = append(validators, vld.SliceUniqueBy(req.AccessibleByProjects,
		func(r *AccessibleByProjectReq) string { return r.ID }))
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAccessibleByProjectsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
