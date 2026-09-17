package apptemplatedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	appNameMaxLen      = 100
	choiceNameMaxLen   = 32
	templateNameMinLen = 1
	// imageRefMaxLen is a registry host, a repository path, a tag and a digest.
	imageRefMaxLen = 512
)

type CreateAppFromTemplateReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`

	Name     string `json:"name"`
	Template string `json:"template"`
	// Version and Variant are empty for the template's defaults.
	Version string `json:"version"`
	Variant string `json:"variant"`
	// ImageOverride replaces the image the template pins, for this app only. It
	// must come from the same repository; the service refuses anything else.
	ImageOverride string `json:"imageOverride"`
	// Params are JSON values of each parameter's type; a size is a string such as "1GB".
	Params map[string]any `json:"params"`
	// DependencyParams are the values a person gave for each dependency the
	// template declares, by the dependency's name: in practice its data volume.
	DependencyParams map[string]map[string]any `json:"dependencyParams"`
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
	req.ImageOverride = strings.TrimSpace(req.ImageOverride)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppFromTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 7) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, appNameMaxLen, "name")...)
	validators = append(validators,
		basedto.ValidateStr(&req.Template, true, templateNameMinLen, templateNameMaxLen, "template")...)
	validators = append(validators, basedto.ValidateStr(&req.Version, false, 1, choiceNameMaxLen, "version")...)
	validators = append(validators, basedto.ValidateStr(&req.Variant, false, 1, choiceNameMaxLen, "variant")...)
	validators = append(validators,
		basedto.ValidateStr(&req.ImageOverride, false, 1, imageRefMaxLen, "imageOverride")...)
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
}

type CreatedDependencyResp struct {
	// Name is the dependency's role in the template: db, cache.
	Name       string                `json:"name"`
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}
