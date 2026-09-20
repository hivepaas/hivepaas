package apptemplatedto

import (
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

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
	Source      string                    `json:"source"`
	Revision    string                    `json:"revision"`
	Name        string                    `json:"name"`
	Title       string                    `json:"title"`
	Tagline     string                    `json:"tagline"`
	Description string                    `json:"description"`
	Categories  []string                  `json:"categories"`
	Tags        []string                  `json:"tags"`
	IconURL     string                    `json:"iconUrl"`
	Links       *AppTemplateLinksResp     `json:"links"`
	License     string                    `json:"license"`
	Compatible  bool                      `json:"compatible"`
	Variants    []*AppTemplateVariantResp `json:"variants"`
	Versions    []*AppTemplateVersionResp `json:"versions"`
	Parameters  []*AppTemplateParamResp   `json:"parameters"`
	// Dependencies are the apps creating this template also creates, each with the
	// parameters a person has to fill in for it.
	Dependencies []*AppTemplateDependencyResp `json:"dependencies"`
	// Capabilities is what creating this template grants the app beyond what a
	// container ordinarily gets, and is null for the templates that ask for
	// nothing. It is shown before anybody deploys, and creating one needs Write
	// on the cluster module.
	Capabilities *AppTemplateCapabilitiesResp `json:"capabilities"`
	// PublishedPorts are the addresses this app claims on the cluster itself,
	// beside the web addresses the reverse proxy serves. Two apps cannot share
	// one, so they are shown before anybody deploys.
	PublishedPorts []*AppTemplatePortResp `json:"publishedPorts"`
}

// AppTemplatePortResp is one port a template publishes on every node.
type AppTemplatePortResp struct {
	// Target is the port inside the container, Published the one on the nodes.
	Target    uint32 `json:"target"`
	Published uint32 `json:"published"`
	// PublishedParam names the parameter that chooses the published port, when
	// one does; Published is then that parameter's default, and what a person
	// types for it is what will actually be claimed.
	PublishedParam string `json:"publishedParam,omitempty"`
	// Protocol is tcp, udp or sctp; PublishMode is ingress - the port answers on
	// every node - or host, only on the node running the app.
	Protocol    string `json:"protocol"`
	PublishMode string `json:"publishMode"`
}

// AppTemplateCapabilitiesResp is the capabilities block of the template, as the
// template wrote it. A version cannot override it, so this is what will be
// granted whichever version is chosen.
type AppTemplateCapabilitiesResp struct {
	// CapabilityAdd and CapabilityDrop are named as docker names them, without
	// the CAP_ prefix: NET_ADMIN, SYS_NICE.
	CapabilityAdd  []string                 `json:"capabilityAdd,omitempty"`
	CapabilityDrop []string                 `json:"capabilityDrop,omitempty"`
	Sysctls        map[string]string        `json:"sysctls,omitempty"`
	Ulimits        []*AppTemplateUlimitResp `json:"ulimits,omitempty"`
	EnableGPU      bool                     `json:"enableGPU,omitempty"`
	OomScoreAdj    int64                    `json:"oomScoreAdj,omitempty"`
}

type AppTemplateUlimitResp struct {
	Name string `json:"name"`
	Soft int64  `json:"soft"`
	Hard int64  `json:"hard"`
}

type AppTemplateDependencyResp struct {
	Name          string `json:"name"`
	Title         string `json:"title"`
	Template      string `json:"template"`
	TemplateTitle string `json:"templateTitle"`
	// Version and Variant are empty for the dependency template's defaults.
	Version string `json:"version"`
	Variant string `json:"variant"`
	// Parameters are only those the person is asked: the template fixes the rest,
	// or they have defaults, or HivePaaS generates them.
	Parameters []*AppTemplateParamResp `json:"parameters"`
	// Capabilities is what this dependency's app is granted. It is gated on the
	// same permission as the main app's: both are created by the one request.
	Capabilities *AppTemplateCapabilitiesResp `json:"capabilities"`
	// PublishedPorts are the ports this dependency's app claims on the cluster.
	PublishedPorts []*AppTemplatePortResp `json:"publishedPorts"`
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
	// Engine is set on an app parameter: the kind of app it may name, which is
	// what the list offered for it is filtered by.
	Engine string `json:"engine,omitempty"`
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
		Source:      tmpl.Source,
		Revision:    tmpl.Revision,
		Name:        metadata.Name,
		Title:       metadata.Title,
		Tagline:     metadata.Tagline,
		Description: metadata.Description,
		Categories:  metadata.Categories,
		Tags:        metadata.Tags,
		IconURL:     summary.IconURL,
		License:     metadata.License,
		Compatible:  summary.Compatible,
		Versions:    summary.Versions,
		Variants:    make([]*AppTemplateVariantResp, 0, len(tmpl.Template.Variants)),
		Parameters:  make([]*AppTemplateParamResp, 0, len(tmpl.Template.Parameters)),
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
	resp.Capabilities = transformCapabilities(tmpl.Template)
	resp.PublishedPorts = transformPublishedPorts(tmpl.Template)
	resp.Dependencies = make([]*AppTemplateDependencyResp, 0, len(tmpl.Dependencies))
	for _, dep := range tmpl.Dependencies {
		asked := dep.Dependency.AskedParams(dep.Template)
		depResp := &AppTemplateDependencyResp{
			Name:          dep.Dependency.Name,
			Title:         dep.Dependency.Title,
			Template:      dep.Dependency.Template,
			TemplateTitle: dep.Template.Metadata.Title,
			Version:       dep.Dependency.Version,
			Variant:       dep.Dependency.Variant,
			Parameters:    make([]*AppTemplateParamResp, 0, len(asked)),
			Capabilities:  transformCapabilities(dep.Template),
		}
		for _, param := range asked {
			depResp.Parameters = append(depResp.Parameters, transformParam(param))
		}
		depResp.PublishedPorts = transformPublishedPorts(dep.Template)
		resp.Dependencies = append(resp.Dependencies, depResp)
	}
	return resp
}

// transformCapabilities reads the block out of the template. A block that
// cannot be read is left out: validation refuses such a template, so rendering
// it would fail before anything was created, and there is nothing truthful to
// show about what it would have granted.
func transformCapabilities(tmpl *templatemodel.Template) *AppTemplateCapabilitiesResp {
	capabilities, err := tmpl.Capabilities()
	if err != nil || capabilities == nil {
		return nil
	}
	resp := &AppTemplateCapabilitiesResp{
		CapabilityAdd:  capabilities.CapabilityAdd,
		CapabilityDrop: capabilities.CapabilityDrop,
		Sysctls:        capabilities.Sysctls,
		EnableGPU:      capabilities.EnableGPU,
		OomScoreAdj:    capabilities.OomScoreAdj,
	}
	for _, ulimit := range capabilities.Ulimits {
		if ulimit == nil {
			continue
		}
		resp.Ulimits = append(resp.Ulimits,
			&AppTemplateUlimitResp{Name: ulimit.Name, Soft: ulimit.Soft, Hard: ulimit.Hard})
	}
	return resp
}

// transformPublishedPorts reads the ports out of the template, filling in what
// docker assumes when the template leaves it out, so that what is shown is what
// will actually be published.
func transformPublishedPorts(tmpl *templatemodel.Template) []*AppTemplatePortResp {
	ports := tmpl.PublishedPorts()
	resp := make([]*AppTemplatePortResp, 0, len(ports))
	for _, port := range ports {
		resp = append(resp, &AppTemplatePortResp{
			Target:         port.Target,
			Published:      port.Published,
			PublishedParam: port.PublishedParam,
			Protocol:       gofn.Coalesce(port.Protocol, string(network.TCP)),
			PublishMode:    gofn.Coalesce(port.PublishMode, string(swarm.PortConfigPublishModeIngress)),
		})
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
		Engine:      param.Engine,
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
