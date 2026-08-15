package projectsettingsdto

import (
	"slices"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

type GetUserAccessesReq struct {
	ProjectID string `json:"-"`
}

func NewGetUserAccessesReq() *GetUserAccessesReq {
	return &GetUserAccessesReq{}
}

func (req *GetUserAccessesReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetUserAccessesResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *UserAccessesDataResp `json:"data"`
}

type UserAccessesDataResp struct {
	OwnerAccess        *UserAccessResp         `json:"ownerAccess"`
	UserAccesses       []*UserAccessResp       `json:"userAccesses"`
	EnvUserAccesses    []*EnvUserAccessesResp  `json:"envUserAccesses"`
	ModuleUserAccesses []*ModuleUserAccessResp `json:"moduleUserAccesses"`
	CurrentUserActions *CurrentUserActionsResp `json:"currentUserActions"`
	UpdateVer          int                     `json:"updateVer"`
}

type UserAccessResp struct {
	*basedto.UserBaseResp
	Access base.AccessActions `json:"access"`
}

type EnvUserAccessesResp struct {
	Name         string            `json:"name"`
	UserAccesses []*UserAccessResp `json:"userAccesses"`
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
	Project            *entity.Project
	ModulePermissions  []*entity.ACLPermission
	ProjectPermissions []*entity.ACLPermission
	EnvPermissions     map[string][]*entity.ACLPermission
	CurrentUser        *entity.User
}

func TransformUserAccesses(input *UserAccessesTransformInput) *UserAccessesDataResp {
	resp := &UserAccessesDataResp{
		OwnerAccess:        TransformOwnerAccessOnProject(input),
		UserAccesses:       TransformUserAccessesOnProject(input),
		EnvUserAccesses:    TransformUserAccessesOnProjectEnvs(input),
		ModuleUserAccesses: TransformUserAccessesOnModule(input),
		UpdateVer:          input.Project.UpdateVer,
	}
	TransformCurrentUserActions(input, resp)
	return resp
}

func TransformOwnerAccessOnProject(input *UserAccessesTransformInput) *UserAccessResp {
	return &UserAccessResp{
		UserBaseResp: projectdto.TransformProjectOwner(input.Project),
		Access:       base.NewFullAccessActions(),
	}
}

func TransformUserAccessesOnProject(input *UserAccessesTransformInput) []*UserAccessResp {
	perms := input.ProjectPermissions
	slices.SortStableFunc(perms, func(a, b *entity.ACLPermission) int {
		return strings.Compare(a.SubjectUser.FullName, b.SubjectUser.FullName)
	})

	resp := make([]*UserAccessResp, 0, len(perms))
	for _, access := range perms {
		resp = append(resp, &UserAccessResp{
			UserBaseResp: basedto.TransformUserBase(access.SubjectUser),
			Access:       access.Actions,
		})
	}
	return resp
}

func TransformUserAccessesOnProjectEnvs(input *UserAccessesTransformInput) (resp []*EnvUserAccessesResp) {
	resp = make([]*EnvUserAccessesResp, 0, len(input.Project.ProjectEnvs))
	for _, env := range input.Project.ProjectEnvs {
		perms := input.EnvPermissions[env.ID]
		slices.SortStableFunc(perms, func(a, b *entity.ACLPermission) int {
			return strings.Compare(a.SubjectUser.FullName, b.SubjectUser.FullName)
		})

		envResp := make([]*UserAccessResp, 0, len(perms))
		for _, access := range perms {
			envResp = append(envResp, &UserAccessResp{
				UserBaseResp: basedto.TransformUserBase(access.SubjectUser),
				Access:       access.Actions,
			})
		}
		resp = append(resp, &EnvUserAccessesResp{
			Name:         env.Name,
			UserAccesses: envResp,
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
