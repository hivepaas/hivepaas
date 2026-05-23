package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type ListUserReq struct {
	Status []base.UserStatus `json:"-" mapstructure:"status"`
	Role   []base.UserRole   `json:"-" mapstructure:"role"`
	Search string            `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListUserReq() *ListUserReq {
	return &ListUserReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "full_name"}},
		},
	}
}

func (req *ListUserReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateSlice(req.Status, true, 0,
		base.AllUserStatuses, "status")...)
	validators = append(validators, basedto.ValidateSlice(req.Role, true, 0,
		base.AllUserRoles, "role")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListUserResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*UserResp       `json:"data"`
}

func TransformUser(user *entity.User) (resp *UserResp, err error) {
	if err = copier.Copy(&resp, user); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.MfaTotpActivated = user.TotpSecret != ""
	return resp, nil
}

func TransformUsers(users []*entity.User) ([]*UserResp, error) {
	return basedto.TransformObjectSlice(users, TransformUser) //nolint:wrapcheck
}
