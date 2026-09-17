package apptemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const templateNameMaxLen = 63

type GetAppTemplateReq struct {
	Name string `json:"-"`
}

func NewGetAppTemplateReq() *GetAppTemplateReq {
	return &GetAppTemplateReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, templateNameMaxLen, "name")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppTemplateResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *AppTemplateResp `json:"data"`
}

type AppTemplateResp struct {
	Source              string                    `json:"source"`
	Revision            string                    `json:"revision"`
	Name                string                    `json:"name"`
	Title               string                    `json:"title"`
	Tagline             string                    `json:"tagline"`
	Description         string                    `json:"description"`
	Categories          []string                  `json:"categories"`
	Tags                []string                  `json:"tags"`
	IconURL             string                    `json:"iconUrl"`
	Links               *AppTemplateLinksResp     `json:"links"`
	License             string                    `json:"license"`
	Compatible          bool                      `json:"compatible"`
	RequiresVersionCode string                    `json:"requiresVersionCode"`
	Variants            []*AppTemplateVariantResp `json:"variants"`
	Versions            []*AppTemplateVersionResp `json:"versions"`
	Parameters          []*AppTemplateParamResp   `json:"parameters"`
}

type AppTemplateLinksResp struct {
	Website       string `json:"website,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	Source        string `json:"source,omitempty"`
}

type AppTemplateVariantResp struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

type AppTemplateParamResp struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	// Default is never set for a secret.
	Default   any    `json:"default,omitempty"`
	Optional  bool   `json:"optional"`
	Pattern   string `json:"pattern,omitempty"`
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Min       any    `json:"min,omitempty"`
	Max       any    `json:"max,omitempty"`
	// Generated says a secret left empty is generated.
	Generated bool                          `json:"generated"`
	Options   []*AppTemplateParamOptionResp `json:"options,omitempty"`
}

type AppTemplateParamOptionResp struct {
	Value string `json:"value"`
	Title string `json:"title"`
}

func TransformAppTemplate(tmpl *apptemplateservice.TemplateResp, currentVersionCode string) *AppTemplateResp {
	metadata := tmpl.Template.Metadata
	summary := transformSummary(tmpl.Entry, currentVersionCode)
	resp := &AppTemplateResp{
		Source:              tmpl.Source,
		Revision:            tmpl.Revision,
		Name:                metadata.Name,
		Title:               metadata.Title,
		Tagline:             metadata.Tagline,
		Description:         metadata.Description,
		Categories:          metadata.Categories,
		Tags:                metadata.Tags,
		IconURL:             summary.IconURL,
		License:             metadata.License,
		Compatible:          summary.Compatible,
		RequiresVersionCode: summary.RequiresVersionCode,
		Versions:            summary.Versions,
		Variants:            make([]*AppTemplateVariantResp, 0, len(tmpl.Template.Variants)),
		Parameters:          make([]*AppTemplateParamResp, 0, len(tmpl.Template.Parameters)),
	}
	if links := metadata.Links; links != nil {
		resp.Links = &AppTemplateLinksResp{
			Website: links.Website, Documentation: links.Documentation, Source: links.Source,
		}
	}
	for _, variant := range tmpl.Template.Variants {
		resp.Variants = append(resp.Variants, &AppTemplateVariantResp{
			Name: variant.Name, Title: variant.Title, Description: variant.Description, Default: variant.Default,
		})
	}
	for _, param := range tmpl.Template.Parameters {
		resp.Parameters = append(resp.Parameters, transformParam(param))
	}
	return resp
}

func transformParam(param *templatemodel.Parameter) *AppTemplateParamResp {
	resp := &AppTemplateParamResp{
		Name:        param.Name,
		Title:       param.Title,
		Description: param.Description,
		Type:        string(param.Type),
		Default:     param.Default,
		Optional:    param.Optional,
		Pattern:     param.Pattern,
		MinLength:   param.MinLength,
		MaxLength:   param.MaxLength,
		Min:         param.Min,
		Max:         param.Max,
		Generated:   param.Generate != nil,
	}
	if param.Type == templatemodel.ParamTypeSecret {
		// Validation refuses a secret default already; this keeps one from
		// reaching a response if a template ever slips past it.
		resp.Default = nil
	}
	for _, option := range param.Options {
		resp.Options = append(resp.Options, &AppTemplateParamOptionResp{Value: option.Value, Title: option.Title})
	}
	return resp
}
