package apptemplatedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	appNameMaxLen      = 100
	choiceNameMaxLen   = 32
	templateNameMinLen = 1
	// imageTagMaxLen is docker's own limit on a tag.
	imageTagMaxLen = 128
)

type CreateAppFromTemplateReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`

	Name     string `json:"name"`
	Template string `json:"template"`
	// Version and Variant are empty for the template's defaults.
	Version string `json:"version"`
	Variant string `json:"variant"`
	// ImageTag replaces the tag of the image the template's version pins, for this
	// app only - one of the tags the image-tags endpoint lists. The repository is
	// the template's and is not asked for: everything else the template says is
	// only correct for that software.
	ImageTag string `json:"imageTag"`
	// Params are JSON values of each parameter's type; a size is a string such as "1GB".
	Params map[string]any `json:"params"`
	// DependencyParams are the values a person gave for each dependency the
	// template declares, by the dependency's name: in practice its data volume.
	DependencyParams map[string]map[string]any `json:"dependencyParams"`
	// ResetStorage deletes what a previous install of these apps left in their
	// directories before the new ones are created. It is what the preflight
	// endpoint's findings are answered with, and it is off unless asked for.
	ResetStorage bool `json:"resetStorage,omitempty"`
}

func NewCreateAppFromTemplateReq() *CreateAppFromTemplateReq {
	return &CreateAppFromTemplateReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *CreateAppFromTemplateReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Template = strings.TrimSpace(req.Template)
	req.Version = strings.TrimSpace(req.Version)
	req.Variant = strings.TrimSpace(req.Variant)
	req.ImageTag = strings.TrimSpace(req.ImageTag)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppFromTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 8) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, appNameMaxLen, "name")...)
	validators = append(validators,
		basedto.ValidateStr(&req.Template, true, templateNameMinLen, templateNameMaxLen, "template")...)
	validators = append(validators, basedto.ValidateStr(&req.Version, false, 1, choiceNameMaxLen, "version")...)
	validators = append(validators, basedto.ValidateStr(&req.Variant, false, 1, choiceNameMaxLen, "variant")...)
	validators = append(validators, basedto.ValidateStr(&req.ImageTag, false, 1, imageTagMaxLen, "imageTag")...)
	validators = append(validators,
		basedto.ValidateCond(req.ImageTag == "" || templatemodel.ValidImageTag(req.ImageTag), "imageTag")...)
	validators = append(validators,
		basedto.ValidateCond(len(req.DependencyParams) <= templatemodel.MaxDependencies, "dependencyParams")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAppFromTemplateResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *CreateAppFromTemplateDataResp `json:"data"`
}

type CreateAppFromTemplateDataResp struct {
	App *basedto.ObjectIDResp `json:"app"`
	// Deployment is the app's first deployment, already queued: it pulls the
	// template's image.
	Deployment *basedto.ObjectIDResp `json:"deployment"`
	// Dependencies are the apps created alongside, each with its first deployment.
	Dependencies []*CreatedDependencyResp `json:"dependencies"`
	// Components are the other apps of an application that is several processes,
	// each with its first deployment. The app above is the primary component, so
	// it is not repeated here.
	Components []*CreatedComponentResp `json:"components"`
}

type CreatedComponentResp struct {
	// Name is the component's role in the template: auth, worker.
	Name       string                `json:"name"`
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}

type CreatedDependencyResp struct {
	// Name is the dependency's role in the template: db, cache.
	Name       string                `json:"name"`
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}
