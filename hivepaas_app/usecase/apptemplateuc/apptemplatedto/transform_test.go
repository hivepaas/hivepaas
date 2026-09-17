package apptemplatedto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
		Name: "postgres",
		File: templatemodel.FileRef{Path: "templates/postgres.yaml", SHA256: "file-sha"},
		Icon: templatemodel.FileRef{Path: "icons/postgres.svg",
			SHA256: "3a18fec8536075187ba89eabce24855f2d2f4861ab790daf49cda264cb8ec137"},
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
	index := &apptemplateservice.IndexResp{
		Source:   "official",
		Revision: "abc",
		Index: &templatemodel.Index{
			Categories: []*templatemodel.Category{{ID: "databases", Title: "Databases",
				Children: []*templatemodel.Category{{ID: "sql", Title: "SQL"}}}},
			Tags:      []*templatemodel.Tag{{ID: "sql", Title: "SQL"}},
			Templates: []*templatemodel.IndexEntry{testEntry("v000001")},
		},
	}

	resp := TransformAppTemplateCatalog(index)

	assert.Equal(t, &AppTemplateCatalogResp{
		Source:   "official",
		Revision: "abc",
		Categories: []*AppTemplateCategoryResp{{ID: "databases", Title: "Databases",
			Children: []*AppTemplateCategoryResp{{ID: "sql", Title: "SQL", Children: []*AppTemplateCategoryResp{}}}}},
		Tags: []*AppTemplateTagResp{{ID: "sql", Title: "SQL"}},
	}, resp, "the catalog carries no templates: those are listed a page at a time")
}

func TestTransformAppTemplateSummaries(t *testing.T) {
	useBasePath(t)

	resp := TransformAppTemplateSummaries(
		[]*templatemodel.IndexEntry{testEntry("v000001"), testEntry("v999999")}, "v000001")

	summary := resp[0]
	assert.Equal(t, "/api/app-templates/icons/postgres.3a18fec853.svg", summary.IconURL,
		"the name says which template, the hash piece changes whenever the icon does")
	assert.True(t, summary.Compatible)
	assert.Equal(t, []string{"alpine"}, summary.Versions[0].Variants)
	assert.False(t, resp[1].Compatible, "a template needing a newer HivePaaS is listed but locked")
	assert.Equal(t, "v999999", resp[1].RequiresVersionCode)
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

func TestTransformAppTemplateImageTags(t *testing.T) {
	resp := TransformAppTemplateImageTags(&apptemplateservice.ImageTagsResp{
		Repository: "registry-1.docker.io/library/postgres",
		CurrentTag: "18.6-alpine3.24",
		Truncated:  true,
		Tags: []*apptemplateservice.ImageTag{
			{Tag: "18.7-alpine3.24", Class: templatemodel.ImageOverrideSameLine, Newer: true},
			{Tag: "19.0-alpine3.24", Class: templatemodel.ImageOverrideOtherMajor, Newer: true},
		},
	})

	assert.Equal(t, &AppTemplateImageTagsResp{
		Repository: "registry-1.docker.io/library/postgres",
		CurrentTag: "18.6-alpine3.24",
		Truncated:  true,
		Tags: []*AppTemplateImageTagResp{
			{Tag: "18.7-alpine3.24", Image: "registry-1.docker.io/library/postgres:18.7-alpine3.24",
				Class: "same-line", Newer: true},
			{Tag: "19.0-alpine3.24", Image: "registry-1.docker.io/library/postgres:19.0-alpine3.24",
				Class: "other-major", Newer: true},
		},
	}, resp, "the dashboard posts Image back as imageOverride, so it never builds a reference itself")
}

func TestParseAppTemplateIconFile(t *testing.T) {
	req := NewGetAppTemplateIconReq()
	req.File = "postgres.3a18fec853.svg"
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
	assert.Equal(t, "postgres", req.Name)
	assert.Equal(t, "3a18fec853", req.SHA256Prefix)
	assert.Equal(t, "svg", req.Ext)

	for _, file := range []string{
		"postgres.svg",               // no hash
		"postgres.3a18fec8536.svg",   // hash too long
		"postgres.3A18FEC853.svg",    // hash not lowercase
		"postgres.3a18fec853.jpg",    // not an icon format
		"Postgres.3a18fec853.svg",    // not a template name
		"../postgres.3a18fec853.svg", // not a file name
		"3a18fec8536075187ba89eabce24855f2d2f4861ab790daf49cda264cb8ec137", // the old URL
	} {
		bad := NewGetAppTemplateIconReq()
		bad.File = file
		assert.NoError(t, bad.ModifyRequest())
		assert.NotEmpty(t, bad.Validate(), file)
	}
}

func TestTransformAppTemplateSummaryListsDependencies(t *testing.T) {
	useBasePath(t)
	entry := testEntry("v000001")
	entry.Dependencies = []*templatemodel.IndexDependency{{Name: "db", Title: "Database", Template: "mysql"}}

	summary := transformSummary(entry, "v000001")

	assert.Equal(t, []*AppTemplateDependencySummaryResp{{Name: "db", Title: "Database", Template: "mysql"}},
		summary.Dependencies)
}

func TestTransformAppTemplateAsksOnlyWhatADependencyNeeds(t *testing.T) {
	useBasePath(t)
	mysql := &templatemodel.Template{
		Metadata: templatemodel.Metadata{Name: "mysql", Title: "MySQL"},
		Parameters: []*templatemodel.Parameter{
			{Name: "dbName", Title: "Database name", Type: templatemodel.ParamTypeString, Default: "app"},
			{Name: "password", Title: "Password", Type: templatemodel.ParamTypeSecret,
				Generate: &templatemodel.Generate{Length: 32}},
			{Name: "dataVolume", Title: "Data volume", Type: templatemodel.ParamTypeVolume},
		},
	}
	dep := &templatemodel.Dependency{Name: "db", Title: "Database", Template: "mysql", Version: "8.4"}
	tmpl := &apptemplateservice.TemplateResp{
		Entry: testEntry("v000001"),
		Template: &templatemodel.Template{
			Metadata:     templatemodel.Metadata{Name: "wordpress"},
			Dependencies: []*templatemodel.Dependency{dep},
		},
		Dependencies: []*apptemplateservice.DependencyTemplate{
			{Dependency: dep, Entry: testEntry("v000001"), Template: mysql},
		},
	}

	resp := TransformAppTemplate(tmpl, "v000001")

	assert.Len(t, resp.Dependencies, 1)
	db := resp.Dependencies[0]
	assert.Equal(t, "MySQL", db.TemplateTitle)
	assert.Equal(t, "8.4", db.Version)
	assert.Len(t, db.Parameters, 1)
	assert.Equal(t, "dataVolume", db.Parameters[0].Name, "defaulted and generated parameters are not asked")
}

func TestTransformAppTemplateBindingLinks(t *testing.T) {
	resp := TransformAppTemplateBinding(&entity.AppTemplateSettings{
		Dependencies:    []entity.AppTemplateDependency{{Name: "db", AppID: "app-db", Template: "mysql"}},
		CreatedForAppID: "app-web",
	})

	assert.Equal(t, []*AppTemplateBindingDependencyResp{{Name: "db", AppID: "app-db", Template: "mysql"}},
		resp.Dependencies)
	assert.Equal(t, "app-web", resp.CreatedForAppID)
}
