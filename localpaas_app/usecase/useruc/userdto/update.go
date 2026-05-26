package userdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
)

type UpdateUserReq struct {
	ID              string                       `json:"-"`
	Username        string                       `json:"username"`
	Email           string                       `json:"email"`
	FullName        string                       `json:"fullName"`
	Position        *string                      `json:"position"`
	Status          *base.UserStatus             `json:"status"`
	Role            *base.UserRole               `json:"role"`
	Notes           *string                      `json:"notes"`
	SecurityOption  *base.UserSecurityOption     `json:"securityOption"`
	AccessExpireAt  *time.Time                   `json:"accessExpireAt"`
	ModuleAccesses  basedto.ModuleAccessSliceReq `json:"moduleAccesses"`
	ProjectAccesses basedto.ObjectAccessSliceReq `json:"projectAccesses"`
}

func NewUpdateUserReq() *UpdateUserReq {
	return &UpdateUserReq{}
}

func (req *UpdateUserReq) ModifyRequest() error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strutil.NormalizeEmail(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)
	return nil
}

func (req *UpdateUserReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, validateUsername(&req.Username, false, "username")...)
	validators = append(validators, basedto.ValidateEmail(&req.Email, false, "email")...)
	validators = append(validators, basedto.ValidateStr(&req.FullName, false, nameMinLen, nameMaxLen,
		"fullName")...)
	validators = append(validators, basedto.ValidateStr(req.Position, false,
		nameMinLen, nameMaxLen, "position")...)
	validators = append(validators, basedto.ValidateStrIn(req.Status, false, base.AllUserStatuses,
		"status")...)
	validators = append(validators, basedto.ValidateStrIn(req.Role, false, base.AllUserRoles,
		"role")...)
	validators = append(validators, basedto.ValidateStr(req.Notes, false,
		notesMinLen, notesMaxLen, "notes")...)
	validators = append(validators, basedto.ValidateStrIn(req.SecurityOption, false,
		base.AllUserSecurityOptions, "securityOption")...)
	validators = append(validators, basedto.ValidateModuleAccessSliceReq(req.ModuleAccesses, true,
		0, base.AllResourceModules, "moduleAccesses")...)
	validators = append(validators, basedto.ValidateObjectAccessSliceReq(req.ProjectAccesses, true,
		0, "projectAccesses")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateUserResp struct {
	Meta *basedto.Meta `json:"meta"`
}
