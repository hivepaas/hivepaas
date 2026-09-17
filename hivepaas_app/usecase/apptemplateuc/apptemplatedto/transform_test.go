package apptemplatedto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

func useBasePath(t *testing.T) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{HTTPServer: config.HTTPServer{BasePath: "/api"}})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func testEntry(versionCode string) *templatemodel.IndexEntry {
	return &templatemodel.IndexEntry{
		Name:       "postgres",
		File:       templatemodel.FileRef{Path: "templates/postgres.yaml", SHA256: "file-sha"},
		Icon:       templatemodel.FileRef{Path: "icons/postgres.svg", SHA256: "icon-sha"},
		Title:      "PostgreSQL",
		Tagline:    "A database",
		Categories: []string{"databases/sql"},
		Variants:   []*templatemodel.IndexVariant{{Name: "alpine", Default: true}},
		Versions: []*templatemodel.IndexVersion{
			{Name: "17", Release: "17.6", Default: true, Variants: []string{"alpine"}},
		},
		Requires: templatemodel.Requires{VersionCode: versionCode},
	}
}

func TestTransformAppTemplateCatalog(t *testing.T) {
	useBasePath(t)
	index := &apptemplateservice.IndexResp{
		Source:   "official",
		Revision: "abc",
		Index: &templatemodel.Index{
			Categories: []*templatemodel.Category{{ID: "databases", Title: "Databases",
				Children: []*templatemodel.Category{{ID: "sql", Title: "SQL"}}}},
			Tags:      []*templatemodel.Tag{{ID: "sql", Title: "SQL"}},
			Templates: []*templatemodel.IndexEntry{testEntry("v000001"), testEntry("v999999")},
		},
	}

	resp := TransformAppTemplateCatalog(index, "v000001")

	assert.Equal(t, "official", resp.Source)
	assert.Equal(t, "SQL", resp.Categories[0].Children[0].Title)
	summary := resp.Templates[0]
	assert.Equal(t, "/api/app-templates/icons/icon-sha", summary.IconURL)
	assert.True(t, summary.Compatible)
	assert.Equal(t, []string{"alpine"}, summary.Versions[0].Variants)
	assert.False(t, resp.Templates[1].Compatible, "a template needing a newer HivePaaS is listed but locked")
	assert.Equal(t, "v999999", resp.Templates[1].RequiresVersionCode)
}

func TestTransformAppTemplateNeverSendsASecretDefault(t *testing.T) {
	useBasePath(t)
	tmpl := &apptemplateservice.TemplateResp{
		Source: "official", Revision: "abc", Entry: testEntry("v000001"),
		Template: &templatemodel.Template{
			Metadata: templatemodel.Metadata{Name: "postgres", Title: "PostgreSQL", Description: "Long text.",
				Links: &templatemodel.Links{Website: "https://www.postgresql.org"}},
			Parameters: []*templatemodel.Parameter{
				{Name: "password", Title: "Password", Type: templatemodel.ParamTypeSecret, Default: "leaked",
					Generate: &templatemodel.Generate{Length: 32}},
				{Name: "memoryLimit", Title: "Memory", Type: templatemodel.ParamTypeSize, Default: "512MB",
					Min: "128MB"},
			},
			Variants: []*templatemodel.Variant{{Name: "alpine", Title: "Alpine", Default: true}},
		},
	}

	resp := TransformAppTemplate(tmpl, "v000001")

	assert.Equal(t, "Long text.", resp.Description)
	assert.Equal(t, "https://www.postgresql.org", resp.Links.Website)
	assert.Nil(t, resp.Parameters[0].Default)
	assert.True(t, resp.Parameters[0].Generated)
	assert.Equal(t, "512MB", resp.Parameters[1].Default)
	assert.Equal(t, "128MB", resp.Parameters[1].Min)
	assert.Equal(t, "Alpine", resp.Variants[0].Title)
	assert.Equal(t, "17.6", resp.Versions[0].Release)
}
