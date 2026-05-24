package projectsettingsdto

import (
	"slices"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type GetUserAccessesReq struct {
	ProjectID string `json:"-"`
}

func NewGetUserAccessesReq() *GetUserAccessesReq {
	return &GetUserAccessesReq{}
}

func (req *GetUserAccessesReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetUserAccessesResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *UserAccessesDataResp `json:"data"`
}

type UserAccessesDataResp struct {
	ProjectUserAccesses []*ProjectUserAccessResp `json:"projectUserAccesses"`
	ModuleUserAccesses  []*ModuleUserAccessResp  `json:"moduleUserAccesses"`
	CurrentUserActions  *CurrentUserActionsResp  `json:"currentUserActions"`
}

type ProjectUserAccessResp struct {
	*basedto.UserBaseResp
	Access base.AccessActions `json:"access"`
}

type ModuleUserAccessResp struct {
	*basedto.UserBaseResp
	Access base.AccessActions `json:"access"`
}

type CurrentUserActionsResp struct {
	CanUpdateProjectUserAccesses bool `json:"canUpdateProjectUserAccesses"`
	CanViewModuleUserAccesses    bool `json:"canViewModuleUserAccesses"`
}

type UserAccessesTransformInput struct {
	Project           *entity.Project
	ObjectPermissions []*entity.ACLPermission
	ModulePermissions []*entity.ACLPermission
	CurrentUser       *entity.User
}

func TransformUserAccesses(input *UserAccessesTransformInput) *UserAccessesDataResp {
	resp := &UserAccessesDataResp{
		ProjectUserAccesses: TransformUserAccessesOnProject(input),
		ModuleUserAccesses:  TransformUserAccessesOnModule(input),
	}
	TransformCurrentUserActions(input, resp)
	return resp
}

func TransformUserAccessesOnProject(input *UserAccessesTransformInput) []*ProjectUserAccessResp {
	perms := input.ObjectPermissions
	slices.SortStableFunc(perms, func(a, b *entity.ACLPermission) int {
		return strings.Compare(a.SubjectUser.FullName, b.SubjectUser.FullName)
	})

	resp := make([]*ProjectUserAccessResp, 0, len(perms))
	for _, access := range perms {
		resp = append(resp, &ProjectUserAccessResp{
			UserBaseResp: basedto.TransformUserBase(access.SubjectUser),
			Access:       access.Actions,
		})
	}
	return resp
}

func TransformUserAccessesOnModule(input *UserAccessesTransformInput) []*ModuleUserAccessResp {
	perms := input.ModulePermissions
	slices.SortStableFunc(perms, func(a, b *entity.ACLPermission) int {
		return strings.Compare(a.SubjectUser.FullName, b.SubjectUser.FullName)
	})

	resp := make([]*ModuleUserAccessResp, 0, len(perms))
	for _, access := range perms {
		resp = append(resp, &ModuleUserAccessResp{
			UserBaseResp: basedto.TransformUserBase(access.SubjectUser),
			Access:       access.Actions,
		})
	}
	return resp
}

func TransformCurrentUserActions(
	input *UserAccessesTransformInput,
	resp *UserAccessesDataResp,
) {
	resp.CurrentUserActions = &CurrentUserActionsResp{}
	// Admin and project owner
	if input.CurrentUser.Role == base.UserRoleAdmin || input.CurrentUser.ID == input.Project.OwnerID {
		resp.CurrentUserActions.CanUpdateProjectUserAccesses = true
		resp.CurrentUserActions.CanViewModuleUserAccesses = true
		return
	}
	for _, userAccess := range resp.ModuleUserAccesses {
		if userAccess.ID != input.CurrentUser.ID {
			continue
		}
		if userAccess.Access.Write {
			resp.CurrentUserActions.CanUpdateProjectUserAccesses = true
		}
		resp.CurrentUserActions.CanViewModuleUserAccesses = true
		return
	}
}
