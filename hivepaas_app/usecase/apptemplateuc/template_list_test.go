package apptemplateuc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func filterTestEntries() []*templatemodel.IndexEntry {
	return []*templatemodel.IndexEntry{
		{Name: "mariadb", Title: "MariaDB", Tagline: "A fork of MySQL",
			Categories: []string{"databases/sql"}, Tags: []string{"sql", "mysql-compatible"}, Aliases: []string{"mysql"}},
		{Name: "postgres", Title: "PostgreSQL", Tagline: "An object-relational database",
			Categories: []string{"databases/sql"}, Tags: []string{"sql", "relational"}, Aliases: []string{"pg", "psql"}},
		{Name: "redis", Title: "Redis", Tagline: "An in-memory data store",
			Categories: []string{"databases/cache"}, Tags: []string{"cache"}},
		{Name: "gitea", Title: "Gitea", Tagline: "Self-hosted Git service",
			Categories: []string{"webapps/dev-tools"}, Tags: []string{"git"}},
	}
}

func names(entries []*templatemodel.IndexEntry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name)
	}
	return out
}

func TestFilterTemplates(t *testing.T) {
	for name, tc := range map[string]struct {
		req  *apptemplatedto.ListAppTemplatesReq
		want []string
	}{
		"no filter": {
			&apptemplatedto.ListAppTemplatesReq{},
			[]string{"mariadb", "postgres", "redis", "gitea"},
		},
		"a parent category matches every child": {
			&apptemplatedto.ListAppTemplatesReq{Categories: []string{"databases"}},
			[]string{"mariadb", "postgres", "redis"},
		},
		"a child category matches exactly": {
			&apptemplatedto.ListAppTemplatesReq{Categories: []string{"databases/cache"}},
			[]string{"redis"},
		},
		"categories are ORed": {
			&apptemplatedto.ListAppTemplatesReq{Categories: []string{"databases/cache", "webapps"}},
			[]string{"redis", "gitea"},
		},
		"a parent prefix is not a parent": {
			&apptemplatedto.ListAppTemplatesReq{Categories: []string{"data"}},
			[]string{},
		},
		"a tag matches exactly": {
			&apptemplatedto.ListAppTemplatesReq{Tags: []string{"mysql-compatible"}},
			[]string{"mariadb"},
		},
		"tags are ORed": {
			&apptemplatedto.ListAppTemplatesReq{Tags: []string{"cache", "git"}},
			[]string{"redis", "gitea"},
		},
		"search finds a hidden alias, ignoring case": {
			&apptemplatedto.ListAppTemplatesReq{Search: "MySQL"},
			[]string{"mariadb"},
		},
		"search reads tags and taglines": {
			&apptemplatedto.ListAppTemplatesReq{Search: "relational"},
			[]string{"postgres"},
		},
		"filters are ANDed": {
			&apptemplatedto.ListAppTemplatesReq{Categories: []string{"databases"}, Tags: []string{"sql"},
				Search: "pg"},
			[]string{"postgres"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, names(filterTemplates(filterTestEntries(), tc.req)))
		})
	}
}

func TestPageTemplates(t *testing.T) {
	entries := filterTestEntries()

	page, meta := pageTemplates(entries, basedto.Paging{Offset: 1, Limit: 2})
	assert.Equal(t, []string{"postgres", "redis"}, names(page))
	assert.Equal(t, &basedto.PagingMeta{Offset: 1, Limit: 2, Total: 4}, meta)

	page, meta = pageTemplates(entries, basedto.Paging{Offset: 3, Limit: 2})
	assert.Equal(t, []string{"gitea"}, names(page), "the last page may be short")
	assert.Equal(t, int64(4), meta.Total)

	page, meta = pageTemplates(entries, basedto.Paging{Offset: 10, Limit: 2})
	assert.Empty(t, page, "an offset past the end is an empty page, not an error")
	assert.Equal(t, int64(4), meta.Total)
}

func TestFilteredTotalIsWhatThePagesCover(t *testing.T) {
	filtered := filterTemplates(filterTestEntries(),
		&apptemplatedto.ListAppTemplatesReq{Categories: []string{"databases"}})

	page, meta := pageTemplates(filtered, basedto.Paging{Limit: 1})

	assert.Len(t, page, 1)
	assert.Equal(t, int64(3), meta.Total, "total counts what matched, not the whole catalog")
}

func sortTestEntries() []*templatemodel.IndexEntry {
	return []*templatemodel.IndexEntry{
		{Name: "redis", Stats: &templatemodel.IndexStats{Added: "2026-09-01", Stars: 70000, StarsGained: 300}},
		{Name: "gitea", Stats: &templatemodel.IndexStats{Added: "2026-10-01", Stars: 50000, StarsGained: 900}},
		// Counted nothing: not on GitHub, and from an index that predates dates.
		{Name: "forgejo"},
		{Name: "mariadb", Stats: &templatemodel.IndexStats{Added: "2026-09-01", Stars: 7000, StarsGained: 300}},
		{Name: "valkey", Stats: &templatemodel.IndexStats{Added: "2026-10-01"}},
	}
}

func TestSortTemplates(t *testing.T) {
	for sort, want := range map[string][]string{
		"":          {"forgejo", "gitea", "mariadb", "redis", "valkey"},
		"name":      {"forgejo", "gitea", "mariadb", "redis", "valkey"},
		"-name":     {"valkey", "redis", "mariadb", "gitea", "forgejo"},
		"-stars":    {"redis", "gitea", "mariadb", "forgejo", "valkey"},
		"stars":     {"mariadb", "gitea", "redis", "forgejo", "valkey"},
		"-trending": {"gitea", "mariadb", "redis", "forgejo", "valkey"},
		"-added":    {"gitea", "valkey", "mariadb", "redis", "forgejo"},
		"added":     {"mariadb", "redis", "gitea", "valkey", "forgejo"},
	} {
		t.Run(sort, func(t *testing.T) {
			var orders basedto.Orders
			if sort != "" {
				direction, column := basedto.DirectionAsc, sort
				if sort[0] == '-' {
					direction, column = basedto.DirectionDesc, sort[1:]
				}
				orders = basedto.Orders{{Direction: direction, ColumnName: column}}
			}
			assert.Equal(t, want, names(sortTemplates(sortTestEntries(), orders)))
		})
	}
}

func TestListAppTemplatesReqTakesOneKnownOrder(t *testing.T) {
	for sort, valid := range map[string]bool{
		"": true, "name": true, "stars": true, "trending": true, "added": true,
		"popularity": false, "stars,name": false,
	} {
		req := &apptemplatedto.ListAppTemplatesReq{}
		for _, column := range strings.Split(sort, ",") {
			if column != "" {
				req.Paging.Sort = append(req.Paging.Sort, &basedto.Order{Direction: basedto.DirectionDesc, ColumnName: column})
			}
		}
		assert.Equal(t, valid, req.Validate() == nil, sort)
	}
}
