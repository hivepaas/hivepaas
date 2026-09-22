package apptemplateuc

import (
	"context"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) ListAppTemplates(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.ListAppTemplatesReq,
) (*apptemplatedto.ListAppTemplatesResp, error) {
	index, err := uc.appTemplateService.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	page, pagingMeta := pageTemplates(filterTemplates(index.Index.Templates, req), req.Paging)
	return &apptemplatedto.ListAppTemplatesResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: apptemplatedto.TransformAppTemplateSummaries(page, base.CurrentVersion),
	}, nil
}

// filterTemplates keeps the templates matching every filter the request sets. The
// catalog is one verified file already in memory - a few hundred entries at most -
// so this is a loop, not a query.
//
// TODO: app templates later - move listing to a table synced from the index once
// custom templates exist, for full-text search over descriptions. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func filterTemplates(
	entries []*templatemodel.IndexEntry,
	req *apptemplatedto.ListAppTemplatesReq,
) []*templatemodel.IndexEntry {
	search := strings.ToLower(req.Search)
	out := make([]*templatemodel.IndexEntry, 0, len(entries))
	for _, entry := range entries {
		// An internal template is not offered on its own: it exists for another
		// template to name, and the store showing it would be offering something
		// nobody can use by itself. It stays fetchable by name.
		if entry.Internal {
			continue
		}
		if matchesCategories(entry, req.Categories) && matchesTags(entry, req.Tags) &&
			matchesSearch(entry, search) {
			out = append(out, entry)
		}
	}
	return out
}

// matchesCategories reports whether a template is in any wanted category. A wanted
// value without a slash is a parent, and matches every child under it; the slash
// it is joined with keeps `data` from matching `databases/sql`.
func matchesCategories(entry *templatemodel.IndexEntry, wanted []string) bool {
	if len(wanted) == 0 {
		return true
	}
	for _, category := range entry.Categories {
		for _, want := range wanted {
			if category == want || (!strings.Contains(want, "/") && strings.HasPrefix(category, want+"/")) {
				return true
			}
		}
	}
	return false
}

func matchesTags(entry *templatemodel.IndexEntry, wanted []string) bool {
	if len(wanted) == 0 {
		return true
	}
	for _, tag := range entry.Tags {
		if slices.Contains(wanted, tag) {
			return true
		}
	}
	return false
}

// matchesSearch looks in everything a person might type for a template, aliases
// included: they are search terms a template never shows, which is how `mysql`
// finds MariaDB and `pg` finds PostgreSQL. search is lowercase already.
func matchesSearch(entry *templatemodel.IndexEntry, search string) bool {
	if search == "" {
		return true
	}
	for _, field := range slices.Concat([]string{entry.Name, entry.Title, entry.Tagline}, entry.Tags, entry.Aliases) {
		if strings.Contains(strings.ToLower(field), search) {
			return true
		}
	}
	return false
}

// pageTemplates cuts one page out of the filtered templates. Total is what matched,
// so the store can tell how many pages the filter has.
func pageTemplates(
	entries []*templatemodel.IndexEntry,
	paging basedto.Paging,
) ([]*templatemodel.IndexEntry, *basedto.PagingMeta) {
	total := len(entries)
	start := min(paging.Offset, total)
	end := total
	if paging.Limit > 0 {
		end = min(start+paging.Limit, total)
	}
	return entries[start:end], &basedto.PagingMeta{Offset: paging.Offset, Limit: paging.Limit, Total: total}
}
