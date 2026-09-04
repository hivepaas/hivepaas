package projectdto

import (
	"slices"
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
)

type GetProjectReq struct {
	ID              string `json:"-"`
	GetUserAccesses bool   `json:"-" mapstructure:"getUserAccesses"`
}

func NewGetProjectReq() *GetProjectReq {
	return &GetProjectReq{}
}

func (req *GetProjectReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *ProjectResp  `json:"data"`
}

type ProjectResp struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Key          string                   `json:"key"`
	Status       base.ProjectStatus       `json:"status"`
	Photo        string                   `json:"photo"`
	Note         string                   `json:"note"`
	Envs         []*ProjectEnvResp        `json:"envs"`
	Tags         []string                 `json:"tags" copy:"-"` // manual copy ProjectTag -> string
	UserAccesses []*ProjectUserAccessResp `json:"userAccesses"`
	Owner        *basedto.UserBaseResp    `json:"owner"`
	UpdateVer    int                      `json:"updateVer"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProjectAppResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectEnvResp struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Status    base.ProjectStatus `json:"status"`
	Color     string             `json:"color"`
	UpdateVer int                `json:"updateVer"`
}

type ProjectUserAccessResp struct {
	*basedto.UserBaseResp
	Access base.AccessActions `json:"access"`
}

type ProjectBaseResp struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Key    string             `json:"key"`
	Photo  string             `json:"photo"`
	Status base.ProjectStatus `json:"status"`
	// Envs lets a caller listing projects also offer their envs without a second
	// request, which is what per-env permission editing needs.
	Envs []*ProjectEnvResp `json:"envs"`
}

func TransformProject(project *entity.Project) (resp *ProjectResp, err error) {
	if err = copier.Copy(&resp, &project); err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Photo = basedto.TransformObjectIcon(project.Photo)
	resp.Envs, err = TransformProjectEnvs(project.ProjectEnvs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Tags = gofn.MapSlice(project.Tags, func(t *entity.Tag) string { return t.Tag })
	resp.UserAccesses = TransformUserAccesses(project)
	resp.Owner = TransformProjectOwner(project)
	return resp, nil
}

func TransformProjectOwner(project *entity.Project) *basedto.UserBaseResp {
	if project.Owner == nil {
		return basedto.NewMissingUser(project.OwnerID)
	}
	return basedto.TransformUserBase(project.Owner)
}

func TransformProjectEnvs(envs []*entity.ProjectEnv) (resp []*ProjectEnvResp, err error) {
	if err = copier.Copy(&resp, &envs); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func TransformUserAccesses(project *entity.Project) []*ProjectUserAccessResp {
	accesses := project.Accesses
	slices.SortStableFunc(accesses, func(a, b *entity.ACLPermission) int {
		return strings.Compare(a.SubjectUser.FullName, b.SubjectUser.FullName)
	})

	// Add owner to the list if not exist
	if project.Owner != nil {
		var ownerAccess *entity.ACLPermission
		for _, access := range accesses {
			if access.SubjectID == project.OwnerID {
				ownerAccess = access
				break
			}
		}
		if ownerAccess == nil {
			ownerAccess = &entity.ACLPermission{
				SubjectID:    project.OwnerID,
				SubjectType:  base.SubjectTypeUser,
				SubjectUser:  project.Owner,
				ResourceType: base.ResourceTypeProject,
				ResourceID:   project.ID,
			}
			accesses = append([]*entity.ACLPermission{ownerAccess}, accesses...)
		}
		ownerAccess.Actions = base.NewFullAccessActions()
	}

	resp := make([]*ProjectUserAccessResp, 0, len(accesses))
	for _, access := range accesses {
		resp = append(resp, &ProjectUserAccessResp{
			UserBaseResp: basedto.TransformUserBase(access.SubjectUser),
			Access:       access.Actions,
		})
	}
	return resp
}

func TransformProjectsBase(projects []*entity.Project) []*ProjectBaseResp {
	return gofn.MapSlice(projects, TransformProjectBase)
}

func TransformProjectBase(project *entity.Project) *ProjectBaseResp {
	if project == nil {
		return nil
	}
	// The envs of a project are a handful of small rows, and the error can only
	// come from the copier: fall back to no envs rather than failing the listing.
	envs, _ := TransformProjectEnvs(project.ProjectEnvs)
	return &ProjectBaseResp{
		ID:     project.ID,
		Name:   project.Name,
		Key:    project.Key,
		Photo:  basedto.TransformObjectIcon(project.Photo),
		Status: project.Status,
		Envs:   envs,
	}
}
