package templaterender

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const pgTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: pg
  title: PG
  tagline: Test database
  description: Test.
  categories: [databases/sql]
  icon: icons/pg.svg
  requires: {versionCode: v000001}
parameters:
  - {name: username, title: Username, type: string, default: app, pattern: '^[a-z_]+$'}
  - {name: password, title: Password, type: secret, generate: {length: 24}}
  - {name: memoryLimit, title: Memory limit, type: size, default: 512MB, min: 128MB}
  - {name: dataVolume, title: Data volume, type: volume}
variants:
  - {name: alpine, title: Alpine, default: true}
  - {name: debian, title: Debian}
versions:
  - name: "18"
    release: "18.1"
    images: {alpine: "postgres:18.1-alpine3.22", debian: "postgres:18.1-trixie"}
    override:
      app:
        deployment:
          storage:
            mounts:
              /var/lib/postgresql/data: null
              /var/lib/postgresql: {type: volume, source: "${{ params.dataVolume }}"}
  - name: "17"
    release: "17.6"
    default: true
    images: {alpine: "postgres:17.6-alpine3.22", debian: "postgres:17.6-trixie"}
  - name: "16"
    release: "16.10"
    deprecated: true
    images: {alpine: "postgres:16.10-alpine3.22"}
app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        /var/lib/postgresql/data: {type: volume, source: "${{ params.dataVolume }}"}
    container:
      healthcheck: {enabled: true, command: "pg_isready -U ${{ params.username }}", interval: 10s}
    resources:
      limits: {memory: "${{ params.memoryLimit }}"}
  settings:
    kind:
      category: database
      engine: postgres
      version: "${{ version.name }}"
      database: {username: "${{ params.username }}", password: "${{ params.password }}"}
    envVars:
      data:
        - {k: POSTGRES_PASSWORD, v: "${HIVEPAAS_PASSWORD}"}
    routing: {port: 5432}
`

func pgTemplate(t *testing.T) *templatemodel.Template {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(pgTemplateYAML))
	assert.NoError(t, err)
	return tmpl
}

func render(t *testing.T, req *Request) (*Result, error) {
	t.Helper()
	if req.Template == nil {
		req.Template = pgTemplate(t)
	}
	return Render(req)
}

func TestRenderDefaults(t *testing.T) {
	result, err := render(t, &Request{Params: map[string]any{"dataVolume": "vol-1"}})

	assert.NoError(t, err)
	assert.Equal(t, "17", result.Version.Name)
	assert.Equal(t, "alpine", result.Variant.Name)
	assert.Equal(t, "postgres:17.6-alpine3.22", result.Image)

	deployment := result.Doc.Deployment
	assert.Equal(t, map[string]any{"image": "postgres:17.6-alpine3.22"}, deployment.Source["imageSource"])
	assert.Equal(t, "vol-1", deployment.Storage.Mounts["/var/lib/postgresql/data"].Source)
	assert.Equal(t, "pg_isready -U app", deployment.Container.Healthcheck.Command)
	assert.Equal(t, timeutil.Duration(10*time.Second), deployment.Container.Healthcheck.Interval)
	assert.Equal(t, unit.DataSize(512<<20), deployment.Resources.Limits.Memory)

	password := result.Params["password"]
	assert.True(t, password.Generated)
	kind := result.Doc.Settings["kind"].(map[string]any)
	assert.Equal(t, "17", kind["version"])
	assert.Equal(t, password.Text(), kind["database"].(map[string]any)["password"])

	env := result.Doc.Settings["envVars"].(map[string]any)["data"].([]any)[0].(map[string]any)
	assert.Equal(t, "${HIVEPAAS_PASSWORD}", env["v"], "HivePaaS's own references pass through")
}

func TestRenderBaseLeavesSecretsOut(t *testing.T) {
	given, err := render(t, &Request{Params: map[string]any{"dataVolume": "vol-1", "password": "s3cret-value"}})
	assert.NoError(t, err)

	assert.NotContains(t, string(given.Base), "s3cret-value")
	assert.Contains(t, string(given.Base), "${{ params.password }}")
	assert.Contains(t, string(given.Base), "postgres:17.6-alpine3.22")
	sum := sha256.Sum256(given.Base)
	assert.Equal(t, hex.EncodeToString(sum[:]), given.BaseSHA256)

	generated, err := render(t, &Request{Params: map[string]any{"dataVolume": "vol-1"}})
	assert.NoError(t, err)
	assert.Equal(t, given.BaseSHA256, generated.BaseSHA256,
		"the base does not depend on a secret's value, so it is stable across renders")
}

func TestRenderAppliesTheVersionOverride(t *testing.T) {
	tmpl := pgTemplate(t)
	result, err := render(t, &Request{Template: tmpl, Version: "18", Params: map[string]any{"dataVolume": "vol-1"}})

	assert.NoError(t, err)
	mounts := result.Doc.Deployment.Storage.Mounts
	assert.Contains(t, mounts, "/var/lib/postgresql")
	assert.NotContains(t, mounts, "/var/lib/postgresql/data")
	assert.Equal(t, "vol-1", mounts["/var/lib/postgresql"].Source)
	assert.Equal(t, "postgres:18.1-alpine3.22", result.Image)

	templateMounts := tmpl.App["deployment"].(map[string]any)["storage"].(map[string]any)["mounts"]
	assert.Contains(t, templateMounts, "/var/lib/postgresql/data", "rendering does not modify the template")
}

func TestRenderVariants(t *testing.T) {
	result, err := render(t, &Request{Variant: "debian", Params: map[string]any{"dataVolume": "v"}})
	assert.NoError(t, err)
	assert.Equal(t, "postgres:17.6-trixie", result.Image)

	_, err = render(t, &Request{Version: "16", Variant: "debian", AllowDeprecated: true,
		Params: map[string]any{"dataVolume": "v"}})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVariantUnavailable)

	_, err = render(t, &Request{Variant: "musl", Params: map[string]any{"dataVolume": "v"}})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVariantUnavailable)
}

func TestRenderVersions(t *testing.T) {
	_, err := render(t, &Request{Version: "16", Params: map[string]any{"dataVolume": "v"}})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVersionDeprecated)

	result, err := render(t, &Request{Version: "16", AllowDeprecated: true, Params: map[string]any{"dataVolume": "v"}})
	assert.NoError(t, err)
	assert.Equal(t, "postgres:16.10-alpine3.22", result.Image)

	_, err = render(t, &Request{Version: "99", Params: map[string]any{"dataVolume": "v"}})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVersionNotFound)
}

func TestRenderRefusesAnUndefinedPlaceholder(t *testing.T) {
	tmpl := pgTemplate(t)
	tmpl.App["deployment"].(map[string]any)["container"].(map[string]any)["healthcheck"].(map[string]any)["command"] =
		"pg_isready -U ${{ params.nope }}"

	_, err := render(t, &Request{Template: tmpl, Params: map[string]any{"dataVolume": "v"}})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	assert.Contains(t, paramDetail(t, err), "params.nope")
}

func TestRenderRefusesWhatCannotBeBuilt(t *testing.T) {
	tmpl := pgTemplate(t)
	tmpl.App["deployment"].(map[string]any)["networks"] = map[string]any{
		"dnsConfig": map[string]any{"nameservers": []any{"1.1.1.1"}},
	}

	_, err := render(t, &Request{Template: tmpl, Params: map[string]any{"dataVolume": "v"}})

	assert.ErrorIs(t, err, hperrors.ErrSpecBlockUnsupported)
}

func TestRenderRefusesInvalidParameters(t *testing.T) {
	_, err := render(t, &Request{Params: map[string]any{"dataVolume": "v", "memoryLimit": "64MB"}})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)

	_, err = render(t, &Request{})
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid, "dataVolume is required")
}
