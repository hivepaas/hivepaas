package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListUserBaseReq struct {
	Status []base.UserStatus `json:"-" mapstructure:"status"`
	Role   []base.UserRole   `json:"-" mapstructure:"role"`
	Search string            `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListUserBaseReq() *ListUserBaseReq {
	return &ListUserBaseReq{
		Status: []base.UserStatus{base.UserStatusActive},
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "full_name"}},
		},
	}
}

// Validate implements interface basedto.ReqValidator
func (req *ListUserBaseReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllUserStatuses, "status")...)
	validators = append(validators, basedto.ValidateSlice(req.Role, true, 0,
		base.AllUserRoles, "role")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListUserBaseResp struct {
	Meta *basedto.ListMeta       `json:"meta"`
	Data []*basedto.UserBaseResp `json:"data"`
}
