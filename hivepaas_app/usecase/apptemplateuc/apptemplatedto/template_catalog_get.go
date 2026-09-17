package apptemplatedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type GetAppTemplateCatalogReq struct {
}

func NewGetAppTemplateCatalogReq() *GetAppTemplateCatalogReq {
	return &GetAppTemplateCatalogReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateCatalogReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetAppTemplateCatalogResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppTemplateCatalogResp `json:"data"`
}

// AppTemplateCatalogResp is what the store needs once, when it opens: where the
// templates come from, and the categories and tags to filter the list by. The
// templates themselves are listed a page at a time by ListAppTemplates.
type AppTemplateCatalogResp struct {
	Source     string                     `json:"source"`
	Revision   string                     `json:"revision"`
	Categories []*AppTemplateCategoryResp `json:"categories"`
	Tags       []*AppTemplateTagResp      `json:"tags"`
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

func TransformAppTemplateCatalog(index *apptemplateservice.IndexResp) *AppTemplateCatalogResp {
	resp := &AppTemplateCatalogResp{
		Source:     index.Source,
		Revision:   index.Revision,
		Categories: transformCategories(index.Index.Categories),
		Tags:       make([]*AppTemplateTagResp, 0, len(index.Index.Tags)),
	}
	for _, tag := range index.Index.Tags {
		resp.Tags = append(resp.Tags, &AppTemplateTagResp{ID: tag.ID, Title: tag.Title})
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
