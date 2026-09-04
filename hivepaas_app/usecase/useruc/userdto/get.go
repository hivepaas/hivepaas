package userdto

import (
	"cmp"
	"slices"
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetUserReq struct {
	ID          string `json:"-"`
	GetAccesses bool   `json:"-" mapstructure:"getAccesses"`
}

func NewGetUserReq() *GetUserReq {
	return &GetUserReq{}
}

func (req *GetUserReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ID, true, "id")...))
}

type GetUserResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *UserDetailsResp `json:"data"`
}

type UserResp struct {
	ID               string                  `json:"id"`
	Username         string                  `json:"username"`
	Email            string                  `json:"email"`
	Role             base.UserRole           `json:"role"`
	Status           base.UserStatus         `json:"status"`
	FullName         string                  `json:"fullName"`
	Photo            string                  `json:"photo"`
	Position         string                  `json:"position"`
	SecurityOption   base.UserSecurityOption `json:"securityOption"`
	MfaTotpActivated bool                    `json:"mfaTotpActivated,omitempty"`
	Notes            string                  `json:"notes,omitempty"`

	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	AccessExpireAt *time.Time `json:"accessExpireAt" copy:",nilonzero"`
	LastAccess     *time.Time `json:"lastAccess" copy:",nilonzero"`
}

type UserDetailsResp struct {
	*UserResp
	ProjectAccesses []*ProjectAccessResp          `json:"projectAccesses"`
	ModuleAccesses  basedto.ObjectAccessSliceResp `json:"moduleAccesses"`
}

type ProjectAccessResp struct {
	Project     *basedto.NamedObjectResp `json:"project"`
	EnvAccesses []*EnvAccessResp         `json:"envAccesses"`
}

// EnvAccessResp is an object access carrying the env color, so a caller can
// render the env badge without looking the project up again.
type EnvAccessResp struct {
	basedto.NamedObjectResp
	Color  string             `json:"color"`
	Access base.AccessActions `json:"access"`
}

// projectAccessData collects, for one project, everything needed to compute the
// user's effective access to each of its envs.
type projectAccessData struct {
	project *entity.Project
	// projectAccess is the project-level ACL. It acts as the default for every env
	// the user has no env-level ACL for, mirroring what permission.LoadProjectAccesses
	// does with makeAdjustment. Nil when the user has no project-level ACL.
	projectAccess *base.AccessActions
	// envAccesses holds the explicit env-level ACLs, keyed by ProjectEnv.ID.
	envAccesses map[string]base.AccessActions
	// envs holds every known env of the project, keyed by ProjectEnv.ID.
	envs map[string]*entity.ProjectEnv
}

// TransformUserDetails builds the user details response, reporting the user's
// access per project env rather than per project.
//
// Every env of a project the user has any access to is listed, so the caller can
// render (and edit) the full project/env matrix without a second lookup. The
// access reported for an env is the effective one: its own ACL when it has one,
// otherwise the project-level ACL, otherwise no access at all.
func TransformUserDetails(user *entity.User) (resp *UserDetailsResp, err error) {
	userResp, err := TransformUser(user)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp = &UserDetailsResp{
		UserResp: userResp,
	}

	projects := make(map[string]*projectAccessData, len(user.Accesses))
	// projectDataOf returns the accumulator for a project, creating it on first
	// sight, and records every env the given project object carries.
	projectDataOf := func(project *entity.Project) *projectAccessData {
		if project == nil {
			return nil
		}
		data := projects[project.ID]
		if data == nil {
			data = &projectAccessData{
				project:     project,
				envAccesses: make(map[string]base.AccessActions, len(project.ProjectEnvs)),
				envs:        make(map[string]*entity.ProjectEnv, len(project.ProjectEnvs)),
			}
			projects[project.ID] = data
		}
		// Prefer whichever copy of the project actually carries a name/envs: the
		// same project can be reached both from a project ACL and from an env one.
		if data.project.Name == "" && project.Name != "" {
			data.project = project
		}
		for _, env := range project.ProjectEnvs {
			data.envs[env.ID] = env
		}
		return data
	}

	for _, access := range user.Accesses {
		if access.ResourceType == base.ResourceTypeProject {
			if data := projectDataOf(access.ResourceProject); data != nil {
				actions := access.Actions
				data.projectAccess = &actions
			}
			continue
		}
		if access.ResourceType == base.ResourceTypeProjectEnv {
			env := access.ResourceProjectEnv
			if env == nil {
				continue
			}
			project := env.Project
			if project == nil {
				// The project relation was not loaded: keep the env under a
				// project we only know the ID of rather than dropping the access.
				project = &entity.Project{ID: env.ProjectID}
			}
			data := projectDataOf(project)
			if data == nil {
				continue
			}
			data.envs[env.ID] = env
			data.envAccesses[env.ID] = access.Actions
			continue
		}
		if access.ResourceType == base.ResourceTypeModule {
			resp.ModuleAccesses = append(resp.ModuleAccesses, &basedto.ObjectAccessResp{
				NamedObjectResp: basedto.NamedObjectResp{
					ID: access.ResourceID,
				},
				Access: access.Actions,
			})
			continue
		}
	}

	for _, data := range projects {
		resp.ProjectAccesses = append(resp.ProjectAccesses, data.transform())
	}

	// Sort project accesses by project names
	slices.SortStableFunc(resp.ProjectAccesses, func(a, b *ProjectAccessResp) int {
		return strings.Compare(a.Project.Name, b.Project.Name)
	})

	return resp, nil
}

func (data *projectAccessData) transform() *ProjectAccessResp {
	envs := make([]*entity.ProjectEnv, 0, len(data.envs))
	for _, env := range data.envs {
		envs = append(envs, env)
	}
	// Envs keep their configured order, falling back to the name so the output is
	// stable even when Index is unset.
	slices.SortStableFunc(envs, func(a, b *entity.ProjectEnv) int {
		if a.Index != b.Index {
			return cmp.Compare(a.Index, b.Index)
		}
		return strings.Compare(a.Name, b.Name)
	})

	var envAccesses []*EnvAccessResp
	for _, env := range envs {
		actions, hasOwnAccess := data.envAccesses[env.ID]
		if !hasOwnAccess && data.projectAccess != nil {
			actions = *data.projectAccess // inherited from the project level
		}
		envAccesses = append(envAccesses, &EnvAccessResp{
			NamedObjectResp: basedto.NamedObjectResp{
				ID:   env.ID,
				Name: env.Name,
			},
			Color:  env.Color,
			Access: actions,
		})
	}

	return &ProjectAccessResp{
		Project: &basedto.NamedObjectResp{
			ID:   data.project.ID,
			Name: data.project.Name,
		},
		EnvAccesses: envAccesses,
	}
}
