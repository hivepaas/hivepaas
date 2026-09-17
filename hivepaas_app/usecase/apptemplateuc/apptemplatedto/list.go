package apptemplatedto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type ListAppTemplatesReq struct {
	ProjectID string `json:"-"`
}

func NewListAppTemplatesReq() *ListAppTemplatesReq {
	return &ListAppTemplatesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ListAppTemplatesReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ProjectID, true, "projectId")...))
}

type ListAppTemplatesResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppTemplateCatalogResp `json:"data"`
}

type AppTemplateCatalogResp struct {
	Source     string                     `json:"source"`
	Revision   string                     `json:"revision"`
	Categories []*AppTemplateCategoryResp `json:"categories"`
	Tags       []*AppTemplateTagResp      `json:"tags"`
	Templates  []*AppTemplateSummaryResp  `json:"templates"`
}

type AppTemplateCategoryResp struct {
	ID       string                     `json:"id"`
	Title    string                     `json:"title"`
	Children []*AppTemplateCategoryResp `json:"children,omitempty"`
}

type AppTemplateTagResp struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type AppTemplateSummaryResp struct {
	Name       string                           `json:"name"`
	Title      string                           `json:"title"`
	Tagline    string                           `json:"tagline"`
	Categories []string                         `json:"categories"`
	Tags       []string                         `json:"tags"`
	Aliases    []string                         `json:"aliases"`
	IconURL    string                           `json:"iconUrl"`
	Variants   []*AppTemplateVariantSummaryResp `json:"variants"`
	Versions   []*AppTemplateVersionResp        `json:"versions"`
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

func TransformAppTemplateCatalog(
	index *apptemplateservice.IndexResp,
	currentVersionCode string,
) *AppTemplateCatalogResp {
	resp := &AppTemplateCatalogResp{
		Source:     index.Source,
		Revision:   index.Revision,
		Categories: transformCategories(index.Index.Categories),
		Tags:       make([]*AppTemplateTagResp, 0, len(index.Index.Tags)),
		Templates:  make([]*AppTemplateSummaryResp, 0, len(index.Index.Templates)),
	}
	for _, tag := range index.Index.Tags {
		resp.Tags = append(resp.Tags, &AppTemplateTagResp{ID: tag.ID, Title: tag.Title})
	}
	for _, entry := range index.Index.Templates {
		resp.Templates = append(resp.Templates, transformSummary(entry, currentVersionCode))
	}
	return resp
}

func transformCategories(categories []*templatemodel.Category) []*AppTemplateCategoryResp {
	out := make([]*AppTemplateCategoryResp, 0, len(categories))
	for _, category := range categories {
		out = append(out, &AppTemplateCategoryResp{
			ID:       category.ID,
			Title:    category.Title,
			Children: transformCategories(category.Children),
		})
	}
	return out
}

func transformSummary(entry *templatemodel.IndexEntry, currentVersionCode string) *AppTemplateSummaryResp {
	summary := &AppTemplateSummaryResp{
		Name:                entry.Name,
		Title:               entry.Title,
		Tagline:             entry.Tagline,
		Categories:          entry.Categories,
		Tags:                entry.Tags,
		Aliases:             entry.Aliases,
		IconURL:             AppTemplateIconURL(entry.Icon.SHA256),
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
	return summary
}

// AppTemplateIconURL is where the dashboard loads an icon from. The route is
// public, because an <img> cannot send the Authorization header, and addressed by
// hash, so the response can be cached forever.
func AppTemplateIconURL(sha256Hex string) string {
	return fmt.Sprintf("%v/app-templates/icons/%v", config.Current().HTTPServer.BasePath, sha256Hex)
}
