package userdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
)

type InviteUserReq struct {
	Email           string                       `json:"email"`
	Role            base.UserRole                `json:"role"`
	SecurityOption  base.UserSecurityOption      `json:"securityOption"`
	AccessExpireAt  time.Time                    `json:"accessExpireAt"`
	SendInviteEmail bool                         `json:"sendInviteEmail"`
	ModuleAccesses  basedto.ModuleAccessSliceReq `json:"moduleAccesses"`
	ProjectAccesses basedto.ObjectAccessSliceReq `json:"projectAccesses"`
}

func NewInviteUserReq() *InviteUserReq {
	return &InviteUserReq{}
}

func (req *InviteUserReq) ModifyRequest() error {
	req.Email = strutil.NormalizeEmail(req.Email)
	return nil
}

func (req *InviteUserReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(&req.Role, true,
		base.AllUserRoles, "role")...)
	validators = append(validators, basedto.ValidateStrIn(&req.SecurityOption, true,
		base.AllUserSecurityOptions, "securityOption")...)
	validators = append(validators, basedto.ValidateModuleAccessSliceReq(req.ModuleAccesses, true,
		0, base.AllResourceModules, "moduleAccesses")...)
	validators = append(validators, basedto.ValidateObjectAccessSliceReq(req.ProjectAccesses, true,
		0, "projectAccesses")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type InviteUserResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *InviteUserDataResp `json:"data"`
}

type InviteUserDataResp struct {
	InviteLink string `json:"inviteLink"`
}
