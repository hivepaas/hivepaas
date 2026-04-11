package projectsettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type GetProjectReq struct {
	ID string `json:"-"`
}

func NewGetProjectReq() *GetProjectReq {
	return &GetProjectReq{}
}

func (req *GetProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
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
	Tags         []string                 `json:"tags" copy:"-"` // manual copy ProjectTag -> string
	Apps         []*ProjectAppResp        `json:"apps"`
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
}

func TransformProject(project *entity.Project) (resp *ProjectResp, err error) {
	if err = copier.Copy(&resp, &project); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Tags = gofn.MapSlice(project.Tags, func(t *entity.ProjectTag) string { return t.Tag })
	resp.UserAccesses = TransformUserAccesses(project.Accesses)
	resp.Owner = basedto.TransformUserBase(project.Owner)
	return resp, nil
}

func TransformUserAccesses(accesses []*entity.ACLPermission) []*ProjectUserAccessResp {
	return gofn.MapSlice(accesses, TransformUserAccess)
}

func TransformUserAccess(access *entity.ACLPermission) *ProjectUserAccessResp {
	return &ProjectUserAccessResp{
		UserBaseResp: basedto.TransformUserBase(access.SubjectUser),
		Access:       access.Actions,
	}
}

func TransformProjectsBase(projects []*entity.Project) []*ProjectBaseResp {
	return gofn.MapSlice(projects, func(project *entity.Project) *ProjectBaseResp {
		return &ProjectBaseResp{
			ID:     project.ID,
			Name:   project.Name,
			Key:    project.Key,
			Photo:  project.Photo,
			Status: project.Status,
		}
	})
}
