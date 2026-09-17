package apptemplateuc

import (
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
	assert.Equal(t, 4, meta.Total)

	page, meta = pageTemplates(entries, basedto.Paging{Offset: 10, Limit: 2})
	assert.Empty(t, page, "an offset past the end is an empty page, not an error")
	assert.Equal(t, 4, meta.Total)
}

func TestFilteredTotalIsWhatThePagesCover(t *testing.T) {
	filtered := filterTemplates(filterTestEntries(),
		&apptemplatedto.ListAppTemplatesReq{Categories: []string{"databases"}})

	page, meta := pageTemplates(filtered, basedto.Paging{Limit: 1})

	assert.Len(t, page, 1)
	assert.Equal(t, 3, meta.Total, "total counts what matched, not the whole catalog")
}
