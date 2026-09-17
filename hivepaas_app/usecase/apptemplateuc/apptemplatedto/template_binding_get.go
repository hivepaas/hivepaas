package apptemplatedto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetAppTemplateBindingReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppTemplateBindingReq() *GetAppTemplateBindingReq {
	return &GetAppTemplateBindingReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateBindingReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppTemplateBindingResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppTemplateBindingResp `json:"data"`
}

// AppTemplateBindingResp is what an app remembers of its template. Parameters are
// left out: they are inputs already applied to the app, where the settings screens
// show them - secrets masked - through their own endpoints.
//
// TODO: app templates phase 2 - the update status, the diff, applying an update
// and detaching. See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type AppTemplateBindingResp struct {
	Source   string `json:"source"`
	Template string `json:"template"`
	Title    string `json:"title"`
	Version  string `json:"version"`
	Release  string `json:"release"`
	Variant  string `json:"variant"`
	// ImageOverride is empty when the app runs the image the template pinned.
	ImageOverride string    `json:"imageOverride"`
	Revision      string    `json:"revision"`
	AppliedAt     time.Time `json:"appliedAt"`
	// Dependencies are the apps created alongside this one.
	Dependencies []*AppTemplateBindingDependencyResp `json:"dependencies"`
	// CreatedForAppID is the app this one was created to serve, empty when it was
	// created on its own.
	CreatedForAppID string `json:"createdForAppId"`
}

type AppTemplateBindingDependencyResp struct {
	Name     string `json:"name"`
	AppID    string `json:"appId"`
	Template string `json:"template"`
}

func TransformAppTemplateBinding(settings *entity.AppTemplateSettings) *AppTemplateBindingResp {
	resp := &AppTemplateBindingResp{
		Source:          settings.Source,
		Template:        settings.Template,
		Title:           settings.Title,
		Version:         settings.Version,
		Release:         settings.Base.Release,
		Variant:         settings.Variant,
		ImageOverride:   settings.ImageOverride,
		Revision:        settings.Base.Revision,
		AppliedAt:       settings.Base.AppliedAt,
		CreatedForAppID: settings.CreatedForAppID,
		Dependencies:    make([]*AppTemplateBindingDependencyResp, 0, len(settings.Dependencies)),
	}
	for _, dep := range settings.Dependencies {
		resp.Dependencies = append(resp.Dependencies,
			&AppTemplateBindingDependencyResp{Name: dep.Name, AppID: dep.AppID, Template: dep.Template})
	}
	return resp
}
