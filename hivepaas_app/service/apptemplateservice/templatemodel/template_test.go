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

func TestDecodeIndexRefusesAnotherKind(t *testing.T) {
	_, err := DecodeIndex([]byte(`{"apiVersion": "hivepaas.com/v1", "kind": "AppTemplate"}`))
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}

// Every installation reads the index its channel pins, whatever version it
// runs. A field added later must not take an older installation's store down.
func TestDecodeIndexAcceptsFieldsItDoesNotKnow(t *testing.T) {
	index, err := DecodeIndex([]byte(`{
		"apiVersion": "hivepaas.com/v1", "kind": "TemplateIndex", "addedLater": true,
		"categories": [], "tags": [],
		"templates": [{"name": "demo", "somethingNew": {"a": 1},
			"file": {"path": "templates/demo.yaml", "sha256": "aa"},
			"icon": {"path": "icons/demo.svg", "sha256": "bb"},
			"title": "Demo", "tagline": "A demo", "categories": ["databases/sql"],
			"versions": [{"name": "1", "release": "1.0", "default": true, "later": 2}],
			"requires": {"versionCode": "v000001"}}]
	}`))

	assert.NoError(t, err)
	assert.Equal(t, "Demo", index.FindTemplate("demo").Title)
	assert.Empty(t, index.Skipped)
}

// An entry this HivePaaS cannot read costs that entry. A newer index may change
// what an older installation reads in it, and one template it cannot list must
// not take the rest of the store with it.
func TestDecodeIndexSkipsAnEntryItCannotRead(t *testing.T) {
	index, err := DecodeIndex([]byte(`{
		"apiVersion": "hivepaas.com/v1", "kind": "TemplateIndex",
		"categories": [{"id": "databases", "children": [null, {"title": "no id"}, {"id": "sql"}]},
			{"id": ["changed"]}, null],
		"tags": [{"id": "sql"}, {"title": "no id"}],
		"templates": [
			{"name": "good", "file": {"path": "templates/good.yaml", "sha256": "aa"},
				"versions": [{"name": "1"}]},
			{"name": "changed", "file": {"path": "templates/changed.yaml", "sha256": "bb"},
				"versions": {"1": {}}},
			{"file": {"path": "templates/anonymous.yaml", "sha256": "cc"}, "versions": [{"name": "1"}]},
			{"name": "fileless", "versions": [{"name": "1"}]},
			{"name": "versionless", "file": {"path": "templates/versionless.yaml", "sha256": "dd"},
				"versions": [null, {"release": "1.0"}]},
			{"name": "good", "file": {"path": "templates/good-again.yaml", "sha256": "ee"},
				"versions": [{"name": "1"}]},
			null
		]
	}`))

	assert.NoError(t, err)
	assert.Len(t, index.Templates, 1)
	assert.Equal(t, "templates/good.yaml", index.FindTemplate("good").File.Path)
	assert.Len(t, index.Categories, 1)
	assert.Equal(t, []*Category{{ID: "sql"}}, index.Categories[0].Children)
	assert.Equal(t, []*Tag{{ID: "sql"}}, index.Tags)

	skipped := map[string]string{}
	for _, entry := range index.Skipped {
		skipped[entry.List+"/"+entry.Name] = entry.Reason
	}
	assert.Contains(t, skipped["templates/changed"], "cannot unmarshal")
	assert.Equal(t, "it has no name", skipped["templates/"])
	assert.Equal(t, "it names no file", skipped["templates/fileless"])
	assert.Equal(t, "it has no version", skipped["templates/versionless"])
	assert.Equal(t, "another entry has its name", skipped["templates/good"])
	assert.Len(t, index.Skipped, 9, "two categories, a tag and six templates, null among them")
}

// A null in one of an entry's lists is dropped rather than read: the store would
// otherwise dereference it.
func TestDecodeIndexDropsNullItems(t *testing.T) {
	index, err := DecodeIndex([]byte(`{
		"apiVersion": "hivepaas.com/v1", "kind": "TemplateIndex",
		"templates": [{"name": "demo", "file": {"path": "templates/demo.yaml", "sha256": "aa"},
			"variants": [null, {"name": "alpine"}], "dependencies": [null], "components": [null],
			"versions": [null, {"name": "1"}]}]
	}`))

	assert.NoError(t, err)
	entry := index.FindTemplate("demo")
	assert.Equal(t, []*IndexVariant{{Name: "alpine"}}, entry.Variants)
	assert.Empty(t, entry.Dependencies)
	assert.Empty(t, entry.Components)
	assert.Equal(t, []*IndexVersion{{Name: "1"}}, entry.Versions)
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
