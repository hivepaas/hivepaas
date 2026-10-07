package apptemplateuc

import (
	"cmp"
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
	matched := sortTemplates(filterTemplates(index.Index.Templates, req), req.Paging.Sort)
	page, pagingMeta := pageTemplates(matched, req.Paging)
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

// sortTemplates puts the templates in the order asked for, by name when none was.
//
// Ties go by name, so that a page boundary falls in the same place on every
// request: templates with the same stars, or none, would otherwise move between
// pages as they are fetched. And a template the order knows nothing about - no
// stars counted, no day it joined - comes after the ones it does, whichever way
// it runs: the store asked for the most starred, or the oldest, not for the ones
// nobody counted.
func sortTemplates(entries []*templatemodel.IndexEntry, orders basedto.Orders) []*templatemodel.IndexEntry {
	column, desc := apptemplatedto.AppTemplateOrderName, false
	if len(orders) > 0 {
		column, desc = orders[0].ColumnName, orders[0].Direction == basedto.DirectionDesc
	}
	byName := func(a, b *templatemodel.IndexEntry) int { return cmp.Compare(a.Name, b.Name) }
	var key func(entry *templatemodel.IndexEntry) (string, int)
	switch column {
	case apptemplatedto.AppTemplateOrderStars:
		key = func(entry *templatemodel.IndexEntry) (string, int) { return "", statsOf(entry).Stars }
	case apptemplatedto.AppTemplateOrderTrending:
		key = func(entry *templatemodel.IndexEntry) (string, int) { return "", statsOf(entry).StarsGained }
	case apptemplatedto.AppTemplateOrderAdded:
		key = func(entry *templatemodel.IndexEntry) (string, int) { return statsOf(entry).Added, 0 }
	default:
		slices.SortStableFunc(entries, func(a, b *templatemodel.IndexEntry) int {
			if desc {
				return byName(b, a)
			}
			return byName(a, b)
		})
		return entries
	}

	slices.SortStableFunc(entries, func(a, b *templatemodel.IndexEntry) int {
		textA, numA := key(a)
		textB, numB := key(b)
		knownA, knownB := textA != "" || numA != 0, textB != "" || numB != 0
		if knownA != knownB {
			if knownA {
				return -1
			}
			return 1
		}
		order := cmp.Or(cmp.Compare(textA, textB), cmp.Compare(numA, numB))
		if desc {
			order = -order
		}
		return cmp.Or(order, byName(a, b))
	})
	return entries
}

var noStats = &templatemodel.IndexStats{}

func statsOf(entry *templatemodel.IndexEntry) *templatemodel.IndexStats {
	if entry.Stats == nil {
		return noStats
	}
	return entry.Stats
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
	return entries[start:end], &basedto.PagingMeta{Offset: paging.Offset, Limit: paging.Limit, Total: int64(total)}
}
