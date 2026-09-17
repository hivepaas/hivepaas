package templaterepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const webTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: web
  title: Web
  tagline: A web app
  description: Web.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
parameters:
  - {name: siteName, title: Site name, type: string, default: site}
dependencies:
  - {name: db, title: Database, template: demo, version: "2"}
versions:
  - {name: "1", release: "1.0", default: true, image: "web:1.0.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
  settings:
    kind: {category: webapp, engine: web, webapp: {}}
    envVars:
      data:
        - {k: DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
`

func lintWeb(t *testing.T, web string) []string {
	t.Helper()
	_, problems := loadAndLint(t, withFile(validRepoFS(), "templates/web.yaml", web))
	messages := make([]string, 0, len(problems))
	for _, problem := range problems {
		messages = append(messages, problem.String())
	}
	return messages
}

func TestLintAcceptsATemplateWithADependency(t *testing.T) {
	assert.Empty(t, lintWeb(t, webTemplateYAML))
}

func TestLintChecksDependenciesAgainstTheRepository(t *testing.T) {
	cases := map[string]struct {
		web  string
		want string
	}{
		"unknown template": {
			strings.Replace(webTemplateYAML, "template: demo", "template: mysql", 1),
			`template "mysql" is not in this repository`,
		},
		"unknown version": {
			strings.Replace(webTemplateYAML, `version: "2"`, `version: "9"`, 1),
			`template "demo" has no version "9"`,
		},
		"deprecated version": {
			strings.Replace(webTemplateYAML, `version: "2"`, `version: "1"`, 1),
			`version "1" of "demo" is deprecated`,
		},
		"unknown parameter": {
			strings.Replace(webTemplateYAML, `version: "2"}`, `version: "2", params: {size: 1}}`, 1),
			`template "demo" declares no parameter "size"`,
		},
		"not shared": {
			strings.Replace(webTemplateYAML, "HIVEPAAS_PASSWORD", "HIVEPAAS_BUCKET", 1),
			`names nothing dependency "db" shares`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, strings.Join(lintWeb(t, tc.web), "\n"), tc.want)
		})
	}
}

func TestLintRefusesANestedDependency(t *testing.T) {
	nested := strings.Replace(demoTemplateYAML, "parameters:",
		"dependencies:\n  - {name: cache, title: Cache, template: web}\nparameters:", 1)
	fsys := withFile(withFile(validRepoFS(), "templates/web.yaml", webTemplateYAML), "templates/demo.yaml", nested)

	_, problems := loadAndLint(t, fsys)

	all := make([]string, 0, len(problems))
	for _, problem := range problems {
		all = append(all, problem.String())
	}
	assert.Contains(t, strings.Join(all, "\n"), "dependencies go one level deep")
}

func TestBuildIndexListsDependencies(t *testing.T) {
	repo, problems := loadAndLint(t, withFile(validRepoFS(), "templates/web.yaml", webTemplateYAML))
	assert.Empty(t, problems)

	index, err := BuildIndex(repo)

	assert.NoError(t, err)
	assert.Equal(t, []*templatemodel.IndexDependency{{Name: "db", Title: "Database", Template: "demo"}},
		index.FindTemplate("web").Dependencies)
}
