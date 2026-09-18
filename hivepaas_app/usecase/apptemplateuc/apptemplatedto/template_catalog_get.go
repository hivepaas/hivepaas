package apptemplatedto

import (
	"strings"

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
	ID    string `json:"id"`
	Title string `json:"title"`
	// Count is how many templates listing this category the store would show.
	//
	// A parent counts a template once however many of its children the template
	// is listed in, because that is what opening the parent reveals: the filter
	// a parent stands for matches every child under it, and the same template
	// cannot be shown twice.
	Count    int                        `json:"count"`
	Children []*AppTemplateCategoryResp `json:"children,omitempty"`
}

type AppTemplateTagResp struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func TransformAppTemplateCatalog(index *apptemplateservice.IndexResp) *AppTemplateCatalogResp {
	resp := &AppTemplateCatalogResp{
		Source:   index.Source,
		Revision: index.Revision,
		Categories: transformCategories(index.Index.Categories, "",
			countTemplatesPerCategory(index.Index.Templates)),
		Tags: make([]*AppTemplateTagResp, 0, len(index.Index.Tags)),
	}
	for _, tag := range index.Index.Tags {
		resp.Tags = append(resp.Tags, &AppTemplateTagResp{ID: tag.ID, Title: tag.Title})
	}
	return resp
}

// transformCategories walks the tree, building each category's full reference
// from the parents above it: a child is `webapps` and `cms` in the vocabulary
// and `webapps/cms` everywhere a template names it.
func transformCategories(
	categories []*templatemodel.Category,
	parent string,
	counts map[string]int,
) []*AppTemplateCategoryResp {
	out := make([]*AppTemplateCategoryResp, 0, len(categories))
	for _, category := range categories {
		ref := category.ID
		if parent != "" {
			ref = parent + "/" + category.ID
		}
		out = append(out, &AppTemplateCategoryResp{
			ID:       category.ID,
			Title:    category.Title,
			Count:    counts[ref],
			Children: transformCategories(category.Children, ref, counts),
		})
	}
	return out
}

// countTemplatesPerCategory counts the templates of every category, keyed by the
// reference a template names it with, and of every parent above it.
//
// It counts what listing would return for that category rather than the mentions
// in the index, which is why a template in two children of one parent adds one
// to the parent and not two.
func countTemplatesPerCategory(entries []*templatemodel.IndexEntry) map[string]int {
	counts := map[string]int{}
	for _, entry := range entries {
		counted := map[string]bool{}
		for _, ref := range entry.Categories {
			if counted[ref] {
				continue
			}
			counted[ref] = true
			counts[ref]++
			if parent, _, found := strings.Cut(ref, "/"); found && !counted[parent] {
				counted[parent] = true
				counts[parent]++
			}
		}
	}
	return counts
}
