package apptemplatedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	maxFilterValues   = 20
	filterValueMaxLen = 127
	searchMaxLen      = 100
)

// ListAppTemplatesReq lists the templates a project can create apps from, a page
// at a time. The categories and tags to filter by come from GetAppTemplateCatalog.
//
// Ordering is by template name and cannot be changed: the catalog is one verified
// file read into memory, so there is no database to sort by, and a sort parameter
// would be accepted and silently ignored.
type ListAppTemplatesReq struct {
	// Categories match a template in any of them. A parent such as `databases`
	// matches every child under it; `databases/sql` matches exactly.
	Categories []string `json:"-" mapstructure:"category"`
	// Tags match a template carrying any of them.
	Tags []string `json:"-" mapstructure:"tag"`
	// Search matches name, title, tagline, tags and aliases, ignoring case.
	Search string `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListAppTemplatesReq() *ListAppTemplatesReq {
	return &ListAppTemplatesReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *ListAppTemplatesReq) ModifyRequest() error {
	req.Search = strings.TrimSpace(req.Search)
	req.Categories = trimValues(req.Categories)
	req.Tags = trimValues(req.Tags)
	return nil
}

// trimValues drops the empty values `?category=a,,b` leaves behind.
func trimValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// Validate implements interface basedto.ReqValidator
func (req *ListAppTemplatesReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 4+len(req.Categories)+len(req.Tags)) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Search, false, 1, searchMaxLen, "search")...)
	validators = append(validators, validateFilterValues(req.Categories, "category")...)
	validators = append(validators, validateFilterValues(req.Tags, "tag")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

func validateFilterValues(values []string, field string) []vld.Validator {
	validators := []vld.Validator{
		vld.SliceLen(values, 0, maxFilterValues).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
		),
	}
	for i := range values {
		validators = append(validators, basedto.ValidateStr(&values[i], true, 1, filterValueMaxLen, field)...)
	}
	return validators
}

type ListAppTemplatesResp struct {
	Meta *basedto.ListMeta         `json:"meta"`
	Data []*AppTemplateSummaryResp `json:"data"`
}

type AppTemplateDependencySummaryResp struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	Template string `json:"template"`
}

type AppTemplateSummaryResp struct {
	Name       string   `json:"name"`
	Title      string   `json:"title"`
	Tagline    string   `json:"tagline"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Aliases    []string `json:"aliases"`
	IconURL    string   `json:"iconUrl"`
	License    string   `json:"license"`
	// Dependencies are the apps creating this template also creates.
	Dependencies []*AppTemplateDependencySummaryResp `json:"dependencies"`
	Variants     []*AppTemplateVariantSummaryResp    `json:"variants"`
	Versions     []*AppTemplateVersionResp           `json:"versions"`
	// Compatible is false for a template needing a newer HivePaaS: the store
	// lists it, locked, rather than hiding it.
	Compatible          bool   `json:"compatible"`
	RequiresVersionCode string `json:"requiresVersionCode"`
}

type AppTemplateVariantSummaryResp struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type AppTemplateVersionResp struct {
	Name       string   `json:"name"`
	Release    string   `json:"release"`
	Default    bool     `json:"default"`
	Deprecated bool     `json:"deprecated"`
	Variants   []string `json:"variants"`
}

func TransformAppTemplateSummaries(
	entries []*templatemodel.IndexEntry,
	currentVersionCode string,
) []*AppTemplateSummaryResp {
	out := make([]*AppTemplateSummaryResp, 0, len(entries))
	for _, entry := range entries {
		out = append(out, transformSummary(entry, currentVersionCode))
	}
	return out
}

func transformSummary(entry *templatemodel.IndexEntry, currentVersionCode string) *AppTemplateSummaryResp {
	summary := &AppTemplateSummaryResp{
		Name:                entry.Name,
		Title:               entry.Title,
		Tagline:             entry.Tagline,
		License:             entry.License,
		Categories:          entry.Categories,
		Tags:                entry.Tags,
		Aliases:             entry.Aliases,
		IconURL:             AppTemplateIconURL(entry),
		Dependencies:        make([]*AppTemplateDependencySummaryResp, 0, len(entry.Dependencies)),
		Variants:            make([]*AppTemplateVariantSummaryResp, 0, len(entry.Variants)),
		Versions:            make([]*AppTemplateVersionResp, 0, len(entry.Versions)),
		Compatible:          templatemodel.IsCompatible(entry.Requires, currentVersionCode),
		RequiresVersionCode: entry.Requires.VersionCode,
	}
	for _, variant := range entry.Variants {
		summary.Variants = append(summary.Variants,
			&AppTemplateVariantSummaryResp{Name: variant.Name, Default: variant.Default})
	}
	for _, version := range entry.Versions {
		summary.Versions = append(summary.Versions, &AppTemplateVersionResp{
			Name:       version.Name,
			Release:    version.Release,
			Default:    version.Default,
			Deprecated: version.Deprecated,
			Variants:   version.Variants,
		})
	}
	for _, dep := range entry.Dependencies {
		summary.Dependencies = append(summary.Dependencies,
			&AppTemplateDependencySummaryResp{Name: dep.Name, Title: dep.Title, Template: dep.Template})
	}
	return summary
}
