package templaterender

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// detailOf is what a person reads: an hperrors error's Error() is only its code,
// and the explanation is in the built detail.
func detailOf(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

const blogTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: blog
  title: Blog
  tagline: Test web app
  description: Test.
  categories: [webapps/cms]
  icon: icons/blog.svg
  requires: {versionCode: v000001}
parameters:
  - {name: siteName, title: Site name, type: string, default: blog}
  - {name: adminPassword, title: Admin password, type: secret, generate: {length: 16}}
dependencies:
  - name: db
    title: Database
    template: pg
    params: {username: "${{ params.siteName }}"}
versions:
  - {name: "7", release: "7.1", default: true, image: "blog:7.1.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
  settings:
    kind: {category: webapp, engine: blog, webapp: {}}
    envVars:
      data:
        - {k: DB_HOST, v: "${{ deps.db.ref.HIVEPAAS_HOST }}:${{ deps.db.ref.HIVEPAAS_PORT }}"}
        - {k: DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
        - {k: DB_APP, v: "${{ deps.db.key }}"}
    routing: {port: 80}
`

func blogTemplate(t *testing.T) *templatemodel.Template {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(blogTemplateYAML))
	assert.NoError(t, err)
	return tmpl
}

func databaseBinding() map[string]*DepBinding {
	shared := append([]string{}, base.AppCommonSharedEnvVars...)
	return map[string]*DepBinding{"db": {
		AppKey:     "blog_db",
		SharedVars: append(shared, base.AppKindSharedEnvVars(base.AppCategoryDatabase)...),
	}}
}

func envValues(t *testing.T, result *Result) map[string]string {
	t.Helper()
	envVars := result.Doc.Settings["envVars"].(map[string]any)["data"].([]any)
	values := map[string]string{}
	for _, item := range envVars {
		entry := item.(map[string]any)
		values[entry["k"].(string)] = entry["v"].(string)
	}
	return values
}

func TestRenderResolvesDependencyReferences(t *testing.T) {
	result, err := Render(&Request{Template: blogTemplate(t), Deps: databaseBinding()})

	assert.NoError(t, err)
	values := envValues(t, result)
	assert.Equal(t, "${blog_db.HIVEPAAS_HOST}:${blog_db.HIVEPAAS_PORT}", values["DB_HOST"])
	assert.Equal(t, "${blog_db.HIVEPAAS_PASSWORD}", values["DB_PASSWORD"],
		"an environment reference, resolved at deploy time: the password is never rendered in")
	assert.Equal(t, "blog_db", values["DB_APP"])
}

func TestRenderRefusesWhatADependencyDoesNotShare(t *testing.T) {
	tmpl := blogTemplate(t)
	envVars := tmpl.App["settings"].(map[string]any)["envVars"].(map[string]any)
	envVars["data"] = []any{map[string]any{"k": "BUCKET", "v": "${{ deps.db.ref.HIVEPAAS_BUCKET }}"}}

	_, err := Render(&Request{Template: tmpl, Deps: databaseBinding()})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	assert.Contains(t, detailOf(t, err), `names nothing dependency "db" shares`)
}

func TestRenderRefusesAnUnboundDependency(t *testing.T) {
	_, err := Render(&Request{Template: blogTemplate(t)})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	assert.Contains(t, detailOf(t, err), `dependency "db" is not bound`)
}

func TestRenderUsesResolvedParams(t *testing.T) {
	tmpl := blogTemplate(t)
	resolved, err := ResolveParams(tmpl.Parameters, nil)
	assert.NoError(t, err)

	result, err := Render(&Request{Template: tmpl, ResolvedParams: resolved, Deps: databaseBinding()})

	assert.NoError(t, err)
	assert.Equal(t, resolved["adminPassword"].Text(), result.Params["adminPassword"].Text(),
		"the render uses the secrets its caller already generated")
}

func TestDependencyParams(t *testing.T) {
	blog := blogTemplate(t)
	owner, err := ResolveParams(blog.Parameters, map[string]any{"siteName": "myblog"})
	assert.NoError(t, err)

	params, err := DependencyParams("blog", blog.FindDependency("db"), pgTemplate(t), owner,
		map[string]any{"dataVolume": "vol-2"})

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"username": "myblog", "dataVolume": "vol-2"}, params)
}

func TestDependencyParamsRefusesWhatIsNotAsked(t *testing.T) {
	blog := blogTemplate(t)
	owner, err := ResolveParams(blog.Parameters, nil)
	assert.NoError(t, err)

	_, err = DependencyParams("blog", blog.FindDependency("db"), pgTemplate(t), owner,
		map[string]any{"username": "someone-else"})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid, "the template fixes the username")
}

func TestDependencyParamsRefusesTheOwnersSecret(t *testing.T) {
	blog := blogTemplate(t)
	blog.Dependencies[0].Params = map[string]any{"username": "${{ params.adminPassword }}"}
	owner, err := ResolveParams(blog.Parameters, nil)
	assert.NoError(t, err)

	_, err = DependencyParams("blog", blog.FindDependency("db"), pgTemplate(t), owner, nil)

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid, "a secret belongs to one app")
}

func TestSharedVarsOf(t *testing.T) {
	result, err := render(t, &Request{Params: map[string]any{"dataVolume": "vol-1"}})
	assert.NoError(t, err)

	shared := SharedVarsOf(result.Doc)

	assert.Equal(t, base.AppCategoryDatabase, KindCategory(result.Doc))
	assert.Contains(t, shared, "HIVEPAAS_PASSWORD")
	assert.Contains(t, shared, "HIVEPAAS_HOST")
	assert.NotContains(t, shared, "HIVEPAAS_BUCKET")
}
