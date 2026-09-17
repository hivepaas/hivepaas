package templatemodel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// validTemplateYAML is the smallest template exercising every structural rule:
// variants, a deprecated version, a secret with generate, a size with a bound.
const validTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: demo
  title: Demo
  tagline: A demo app
  description: Demo description.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  links: {website: "https://example.com"}
  requires: {versionCode: v000001}
parameters:
  - {name: password, title: Password, type: secret, generate: {length: 16}}
  - {name: memoryLimit, title: Memory limit, type: size, default: 256MB, min: 128MB}
  - {name: dataVolume, title: Data volume, type: volume}
variants:
  - {name: alpine, title: Alpine, default: true}
  - {name: debian, title: Debian}
versions:
  - name: "2"
    release: "2.1"
    default: true
    images: {alpine: "demo:2.1.0-alpine", debian: "demo:2.1.0"}
  - name: "1"
    release: "1.9"
    deprecated: true
    images: {alpine: "demo:1.9.3-alpine"}
app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
`

func errorDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestDecodeTemplate(t *testing.T) {
	tmpl, err := DecodeTemplate([]byte(validTemplateYAML))
	assert.NoError(t, err)
	assert.Equal(t, "demo", tmpl.Metadata.Name)
	assert.Equal(t, ParamTypeSecret, tmpl.Parameters[0].Type)
	assert.Equal(t, 16, tmpl.Parameters[0].Generate.Length)
	assert.Equal(t, "2", tmpl.DefaultVersion().Name)
	assert.Equal(t, "alpine", tmpl.DefaultVariant().Name)
	assert.Equal(t, "demo:1.9.3-alpine", tmpl.FindVersion("1").ImageFor("alpine"))
	assert.Nil(t, tmpl.FindVersion("3"))
	assert.Contains(t, tmpl.App, "deployment")
}

func TestDecodeTemplateRefusesUnknownFields(t *testing.T) {
	_, err := DecodeTemplate([]byte(validTemplateYAML + "\nsurprise: true\n"))
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	assert.Contains(t, errorDetail(t, err), "surprise")
}

func TestDecodeIndex(t *testing.T) {
	index, err := DecodeIndex([]byte(`{
		"apiVersion": "hivepaas.com/v1", "kind": "TemplateIndex",
		"categories": [{"id": "databases", "title": "Databases"}],
		"tags": [{"id": "sql", "title": "SQL"}],
		"templates": [{
			"name": "demo", "title": "Demo", "tagline": "A demo app", "categories": ["databases/sql"],
			"file": {"path": "templates/demo.yaml", "sha256": "aa"},
			"icon": {"path": "icons/demo.svg", "sha256": "bb"},
			"versions": [{"name": "2", "release": "2.1", "default": true}],
			"requires": {"versionCode": "v000001"}
		}]
	}`))
	assert.NoError(t, err)
	assert.Equal(t, "templates/demo.yaml", index.FindTemplate("demo").File.Path)
	assert.Nil(t, index.FindTemplate("other"))
	assert.Equal(t, "demo", index.FindIcon("bb").Name)
	assert.Nil(t, index.FindIcon("aa"), "a template file is not an icon")
}

func TestDecodeIndexRefusesAnotherKindAndUnknownFields(t *testing.T) {
	_, err := DecodeIndex([]byte(`{"apiVersion": "hivepaas.com/v1", "kind": "AppTemplate"}`))
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)

	_, err = DecodeIndex([]byte(`{"apiVersion": "hivepaas.com/v1", "kind": "TemplateIndex", "extra": 1}`))
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}

func TestDecodeCategoriesAndTags(t *testing.T) {
	categories, err := DecodeCategories([]byte(`
apiVersion: hivepaas.com/v1
kind: TemplateCategories
categories:
  - id: databases
    title: Databases
    children: [{id: sql, title: SQL}]
`))
	assert.NoError(t, err)
	assert.Equal(t, "sql", categories.Categories[0].Children[0].ID)

	tags, err := DecodeTags([]byte("apiVersion: hivepaas.com/v1\nkind: TemplateTags\ntags: [{id: sql, title: SQL}]\n"))
	assert.NoError(t, err)
	assert.Equal(t, "SQL", tags.Tags[0].Title)

	_, err = DecodeTags([]byte("apiVersion: hivepaas.com/v1\nkind: TemplateCategories\ntags: []\n"))
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}

func TestIsCompatible(t *testing.T) {
	assert.True(t, IsCompatible(Requires{VersionCode: "v000001"}, "v000001"))
	assert.True(t, IsCompatible(Requires{VersionCode: "v000001"}, "v000002"))
	assert.False(t, IsCompatible(Requires{VersionCode: "v000002"}, "v000001"))
}
