# App Templates (Phase 1, Backend) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a user create a fully configured app from a curated template - postgres, mariadb - fetched from the signed `hivepaas/app-templates` pin, with the app remembering the template, version, variant, parameters and rendered base that phase 2's updates will merge against.

**Architecture:** Three pure packages hold the format (`templatemodel`), rendering (`templaterender`) and repository tooling (`templaterepo`); `apptemplateservice` fetches and verifies templates through a `Source` interface; rendering produces a `specmodel.AppDoc`, which a new `specservice.BuildApp` turns into settings and a swarm spec; a new `appprovisionservice` - extracted from `appuc.CreateApp` - creates the app; `apptemplateuc` ties it together behind four endpoints. A `tools/apptemplate` CLI lints and indexes the templates repository with the same code HivePaaS runs.

**Tech Stack:** Go 1.27, `gopkg.in/yaml.v3`, `encoding/json`, `crypto/sha256`, `crypto/rand`, `net/http`, Docker Swarm API types (`github.com/moby/moby/api/types/swarm`, `.../mount`, `.../container`), bun via existing repositories, gin handlers, testify.

**Spec:** `docs/superpowers/specs/2026-09-17-app-templates-design.md`

## Global Constraints

- **Layering** (`docs/ARCHITECTURE.md`): `handler → dto → usecase → service → repository → entity → base`. A service never imports anything under `hivepaas_app/usecase/` (DTOs included). A usecase never imports another usecase. Test files may import anything.
- **Pure packages.** `templatemodel`, `templaterender`, `templaterepo` and `specmodel` import no service, repository, database or network code. `tools/apptemplate` depends only on them.
- **Error codes** are declared in the task that first raises them - `tools/errcodelint` refuses a code nothing references - each with its English message: `ERR_APP_TEMPLATE_*` in `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`, `ERR_SPEC_*` in `errors.spec.en.toml`.
- **Before every commit:** `make fmt`, `go build ./...`, `make lint-local` (the whole repo; `make lint` does not work in the devtools container here), `go test ./...`. `make lint-local` runs `tools/errcodelint` and `tools/goroutinelint` too.
- **Style:** 120-character lines, US spelling, no `//nolint:exhaustive` on this repo's enums - list every value (ARCHITECTURE §7). No dynamic `errors.New`/`fmt.Errorf` in non-test code: raise `hperrors` codes.
- **`make gen-swag`** whenever a DTO changes; `docs/openapi/swagger.json` is committed.
- **Deferred work** is marked `// TODO: app templates phase 2|phase 3|later - <what>. See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.`, or `// TODO: spec import - <what>.`
- **Format constants:** `apiVersion: hivepaas.com/v1`; kinds `AppTemplate`, `TemplateIndex`, `TemplateCategories`, `TemplateTags`; placeholders `${{ ref }}`, escaped as `$${{`; tagline at most 80 characters; 1 to 3 categories of the form `parent/child`; at most 8 tags.
- **Limits:** `index.json` 1 MB; a template file 256 KB; an icon 256 KB, `.svg` or `.png`.
- **Parameter types, closed set:** `string`, `secret`, `int`, `size`, `bool`, `select`, `volume`. A `secret` or `volume` has no `default`. Secrets are generated with `crypto/rand`.
- **Supported `AppDoc` subset (phase 1):** `deployment.source` (`activeMethod: image`, `imageSource.image`, `command`, `workingDir`); `deployment.storage.mounts.<target>` (`type: volume`, `source`, `readOnly`, `volumeOptions.subpath`); `deployment.container.healthcheck`; `deployment.resources.{reservations,limits}` (`cpus`, `memory`, and `pids` on limits); `settings.kind`; `settings.envVars`; `settings.routing.port`.
- **Commits:** the steps below commit per task. If the user has not authorized committing in this session, stage the files and ask instead.

---

## File Structure

**New**

| path | responsibility |
|---|---|
| `hivepaas_app/service/apptemplateservice/templatemodel/` | the format: types, strict decoding, structural validation (Task 1) |
| `hivepaas_app/service/specservice/specmodel/buildable.go` | the `AppDoc` subset phase 1 can build (Task 2) |
| `hivepaas_app/service/apptemplateservice/templaterender/` | merge patch and placeholders (Task 3), parameters (Task 4), `Render` (Task 5) |
| `hivepaas_app/service/apptemplateservice/templaterepo/` | load a checkout, lint it, build `index.json` (Task 6) |
| `hivepaas_app/entity/setting_app_template.go` | the `app-template` setting (Task 7) |
| `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mounts.go` | mount building moved out of `appsettingsuc` (Task 8) |
| `hivepaas_app/service/specservice/specserviceimpl/build*.go` | `BuildApp` and its builders (Task 9) |
| `hivepaas_app/service/appprovisionservice/` | `ProvisionApp`, extracted from `appuc.CreateApp` (Task 10) |
| `hivepaas_app/service/apptemplateservice/{service,source}.go` + `apptemplateserviceimpl/` | sources, cache, the service (Task 11) |
| `hivepaas_app/config/app_templates.go` | `HP_TEMPLATES_DIR` (Task 11) |
| `hivepaas_app/usecase/apptemplateuc/` + `apptemplatedto/` | the catalogue (Task 12), creating from a template and the app's binding (Task 13) |
| `hivepaas_app/interface/api/handler/apptemplatehandler/` | transport (Tasks 12-13) |
| `tools/apptemplate/` | `lint`, `index`, `render`, `pin` (Task 14) |
| `../app-templates/` (the other repository) | categories, tags, postgres and mariadb templates, icons, CI (Task 15) |

**Modified**

| path | change |
|---|---|
| `hivepaas_app/hperrors/errors_app_template.go` (new), `errors_spec.go` | error codes, per task |
| `hivepaas_app/base/setting.go`, `entity/setting_spec.go`, `specmodel/singleton.go` | register `app-template` (Task 7) |
| `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go` | call `volumeservice.BuildAppMounts` (Task 8) |
| `hivepaas_app/service/specservice/service.go`, `types.go`, `specserviceimpl/service.go` | `BuildApp` (Task 9) |
| `hivepaas_app/usecase/specuc/export_test.go` | the fake embeds the grown interface (Task 9) |
| `hivepaas_app/usecase/appuc/create.go`, `uc.go` | call `ProvisionApp` (Task 10) |
| `hivepaas_app/config/config.go` | `AppTemplates` section (Task 11) |
| `hivepaas_app/registry/provides.go`, `interface/api/server/router*.go` | wiring and routes (Tasks 10-13) |
| `docs/DEVELOPMENT.md` | authoring templates locally (Task 14) |

---
## Task 1: The template format - `templatemodel`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/template.go`
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/index.go`
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/validate.go`
- Create: `hivepaas_app/hperrors/errors_app_template.go`
- Create: `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`
- Test: `hivepaas_app/service/apptemplateservice/templatemodel/template_test.go`
- Test: `hivepaas_app/service/apptemplateservice/templatemodel/validate_test.go`

**Interfaces:**
- Consumes: `hperrors`, `pkg/imageref.Parse`, `pkg/unit.ParseDataSizeString`.
- Produces:
  - constants `APIVersion`, `KindAppTemplate`, `KindTemplateIndex`, `KindTemplateCategories`, `KindTemplateTags`, `CharsetAlnum`, `CharsetHex`, `MaxTaglineLen`, `MaxCategories`, `MaxTags`
  - `type ParamType string` with `ParamTypeString`, `ParamTypeSecret`, `ParamTypeInt`, `ParamTypeSize`, `ParamTypeBool`, `ParamTypeSelect`, `ParamTypeVolume`, and `AllParamTypes`
  - types `Template`, `Metadata`, `Links`, `Requires`, `Parameter`, `Generate`, `SelectOption`, `Variant`, `Version`, `Override`
  - types `Index`, `IndexEntry`, `IndexVariant`, `IndexVersion`, `FileRef`, `Category`, `Tag`, `Categories`, `Tags`
  - `func DecodeTemplate(data []byte) (*Template, error)`, `func DecodeIndex(data []byte) (*Index, error)`, `func DecodeCategories(data []byte) (*Categories, error)`, `func DecodeTags(data []byte) (*Tags, error)`
  - `func (t *Template) Validate(fileName string) error`
  - `func (t *Template) DefaultVersion() *Version`, `FindVersion(name string) *Version`, `DefaultVariant() *Variant`, `FindVariant(name string) *Variant`, `func (v *Version) ImageFor(variant string) string`
  - `func (p *Parameter) IntBounds() (minValue, maxValue *int64)`, `func (p *Parameter) SizeBounds() (minValue, maxValue *unit.DataSize)`, `func ToInt64(v any) (int64, bool)`
  - `func IsPinnedImage(image string) bool`, `func IsCompatible(requires Requires, currentVersionCode string) bool`
  - `func (i *Index) FindTemplate(name string) *IndexEntry`, `func (i *Index) FindIcon(sha256 string) *IndexEntry`
  - `hperrors.ErrAppTemplateInvalid`

- [ ] **Step 1: Declare the error and its message**

`hivepaas_app/hperrors/errors_app_template.go`:

```go
package hperrors

// Errors for app templates
var (
	ErrAppTemplateInvalid = NewErr(ErrValueInvalid, "ERR_APP_TEMPLATE_INVALID")
)
```

`hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`:

```toml
# Messages for the errors declared in hivepaas_app/hperrors/errors_app_template.go.
# Adding or removing an error there means editing this file.

ERR_APP_TEMPLATE_INVALID = "The app template is invalid"
```

- [ ] **Step 2: Write the failing decoding tests**

`hivepaas_app/service/apptemplateservice/templatemodel/template_test.go`:

```go
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
```

- [ ] **Step 3: Run the tests to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/`
Expected: FAIL - `undefined: DecodeTemplate` (the package has no source yet).

- [ ] **Step 4: Write the types and decoders**

`hivepaas_app/service/apptemplateservice/templatemodel/template.go`:

```go
// Package templatemodel is the app template format: what a template file, the
// category and tag vocabularies and the generated index contain.
//
// It is the durable half of the feature. A template published today is read by
// versions of HivePaaS that do not exist yet, so the shape here changes only with
// APIVersion, and it shares no types with the dashboard DTOs.
package templatemodel

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// APIVersion is the format version of every file in a templates repository. It
// is the spec format's version, because a template's app block is a spec AppDoc.
const APIVersion = "hivepaas.com/v1"

const (
	KindAppTemplate        = "AppTemplate"
	KindTemplateIndex      = "TemplateIndex"
	KindTemplateCategories = "TemplateCategories"
	KindTemplateTags       = "TemplateTags"
)

// Template is one file under templates/.
type Template struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   Metadata     `yaml:"metadata"`
	Parameters []*Parameter `yaml:"parameters,omitempty"`
	Variants   []*Variant   `yaml:"variants,omitempty"`
	Versions   []*Version   `yaml:"versions"`

	// App is a specmodel.AppDoc with placeholders in it. It stays an untyped tree
	// until it is rendered: version overrides and placeholders apply to the tree,
	// and only the result has to decode as an AppDoc.
	App map[string]any `yaml:"app"`
}

type Metadata struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Tagline     string   `yaml:"tagline"`
	Description string   `yaml:"description"`
	Categories  []string `yaml:"categories"`
	Tags        []string `yaml:"tags,omitempty"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Icon        string   `yaml:"icon"`
	Links       *Links   `yaml:"links,omitempty"`
	License     string   `yaml:"license,omitempty"`
	Requires    Requires `yaml:"requires"`
}

type Links struct {
	Website       string `yaml:"website,omitempty" json:"website,omitempty"`
	Documentation string `yaml:"documentation,omitempty" json:"documentation,omitempty"`
	Source        string `yaml:"source,omitempty" json:"source,omitempty"`
}

// Requires names the oldest HivePaaS able to provision a template, by version
// code (base.CurrentVersion). Codes are fixed-width, so they compare as strings.
type Requires struct {
	VersionCode string `yaml:"versionCode" json:"versionCode"`
}

// IsCompatible reports whether a HivePaaS at currentVersionCode can provision a
// template with these requirements.
func IsCompatible(requires Requires, currentVersionCode string) bool {
	return requires.VersionCode <= currentVersionCode
}

type ParamType string

const (
	ParamTypeString ParamType = "string"
	ParamTypeSecret ParamType = "secret"
	ParamTypeInt    ParamType = "int"
	ParamTypeSize   ParamType = "size"
	ParamTypeBool   ParamType = "bool"
	ParamTypeSelect ParamType = "select"
	ParamTypeVolume ParamType = "volume"
)

// AllParamTypes is closed on purpose: the dashboard generates a form field per
// parameter, and a type it does not know is a field it cannot draw.
//
// TODO: app templates phase 3 - more parameter types (ssl-cert, registry-auth,
// domain). See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
var AllParamTypes = []ParamType{ParamTypeString, ParamTypeSecret, ParamTypeInt, ParamTypeSize,
	ParamTypeBool, ParamTypeSelect, ParamTypeVolume}

type Parameter struct {
	Name        string    `yaml:"name"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Type        ParamType `yaml:"type"`
	Default     any       `yaml:"default,omitempty"`
	Optional    bool      `yaml:"optional,omitempty"`

	Pattern   string          `yaml:"pattern,omitempty"`
	MinLength *int            `yaml:"minLength,omitempty"`
	MaxLength *int            `yaml:"maxLength,omitempty"`
	Min       any             `yaml:"min,omitempty"`
	Max       any             `yaml:"max,omitempty"`
	Generate  *Generate       `yaml:"generate,omitempty"`
	Options   []*SelectOption `yaml:"options,omitempty"`
}

const (
	CharsetAlnum = "alnum"
	CharsetHex   = "hex"
)

type Generate struct {
	Length  int    `yaml:"length"`
	Charset string `yaml:"charset,omitempty"`
}

type SelectOption struct {
	Value string `yaml:"value"`
	Title string `yaml:"title"`
}

type Variant struct {
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Default     bool   `yaml:"default,omitempty"`
}

// Version is a major line, pinned to an exact image per variant.
type Version struct {
	Name       string            `yaml:"name"`
	Release    string            `yaml:"release"`
	Default    bool              `yaml:"default,omitempty"`
	Deprecated bool              `yaml:"deprecated,omitempty"`
	Image      string            `yaml:"image,omitempty"`
	Images     map[string]string `yaml:"images,omitempty"`
	Vars       map[string]string `yaml:"vars,omitempty"`
	Override   *Override         `yaml:"override,omitempty"`
}

// Override is what a version changes in the template. It is a struct with one
// field rather than the app tree itself so that a version can later override
// something other than app.
//
// TODO: app templates later - per-version parameter defaults. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type Override struct {
	App map[string]any `yaml:"app,omitempty"`
}

// ImageFor returns this version's image for a variant; variant is empty for a
// template without variants.
func (v *Version) ImageFor(variant string) string {
	if variant == "" {
		return v.Image
	}
	return v.Images[variant]
}

func (t *Template) DefaultVersion() *Version {
	for _, version := range t.Versions {
		if version != nil && version.Default {
			return version
		}
	}
	return nil
}

func (t *Template) FindVersion(name string) *Version {
	for _, version := range t.Versions {
		if version != nil && version.Name == name {
			return version
		}
	}
	return nil
}

func (t *Template) DefaultVariant() *Variant {
	for _, variant := range t.Variants {
		if variant != nil && variant.Default {
			return variant
		}
	}
	return nil
}

func (t *Template) FindVariant(name string) *Variant {
	for _, variant := range t.Variants {
		if variant != nil && variant.Name == name {
			return variant
		}
	}
	return nil
}

// DecodeTemplate parses a template file strictly. An unknown field is an error:
// an older HivePaaS that skipped a field it did not understand would provision
// something other than what the template describes, and nothing would say so.
func DecodeTemplate(data []byte) (*Template, error) {
	tmpl := &Template{}
	if err := decodeYAMLStrict(data, tmpl); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s", err.Error())
	}
	return tmpl, nil
}

func decodeYAMLStrict(data []byte, out any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
```

`hivepaas_app/service/apptemplateservice/templatemodel/index.go`:

```go
package templatemodel

import (
	"bytes"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// Index is index.json: everything the store lists, filters and searches by,
// without the templates themselves.
type Index struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Categories []*Category   `json:"categories"`
	Tags       []*Tag        `json:"tags"`
	Templates  []*IndexEntry `json:"templates"`
}

type FileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type IndexEntry struct {
	Name       string          `json:"name"`
	File       FileRef         `json:"file"`
	Icon       FileRef         `json:"icon"`
	Title      string          `json:"title"`
	Tagline    string          `json:"tagline"`
	Categories []string        `json:"categories"`
	Tags       []string        `json:"tags,omitempty"`
	Aliases    []string        `json:"aliases,omitempty"`
	Variants   []*IndexVariant `json:"variants,omitempty"`
	Versions   []*IndexVersion `json:"versions"`
	Requires   Requires        `json:"requires"`
}

type IndexVariant struct {
	Name    string `json:"name"`
	Default bool   `json:"default,omitempty"`
}

type IndexVersion struct {
	Name       string   `json:"name"`
	Release    string   `json:"release"`
	Default    bool     `json:"default,omitempty"`
	Deprecated bool     `json:"deprecated,omitempty"`
	Variants   []string `json:"variants,omitempty"`
}

type Category struct {
	ID       string      `yaml:"id" json:"id"`
	Title    string      `yaml:"title" json:"title"`
	Children []*Category `yaml:"children,omitempty" json:"children,omitempty"`
}

type Tag struct {
	ID    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title"`
}

// Categories is categories.yaml.
type Categories struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Categories []*Category `yaml:"categories"`
}

// Tags is tags.yaml.
type Tags struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Tags       []*Tag `yaml:"tags"`
}

func (i *Index) FindTemplate(name string) *IndexEntry {
	for _, entry := range i.Templates {
		if entry.Name == name {
			return entry
		}
	}
	return nil
}

// FindIcon finds the template whose icon has this hash. Icons are served by
// hash, and only a hash listed here is served.
func (i *Index) FindIcon(sha256 string) *IndexEntry {
	for _, entry := range i.Templates {
		if entry.Icon.SHA256 == sha256 {
			return entry
		}
	}
	return nil
}

func DecodeIndex(data []byte) (*Index, error) {
	index := &Index{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(index); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("index.json: %s", err.Error())
	}
	if err := checkHeader("index.json", index.APIVersion, index.Kind, KindTemplateIndex); err != nil {
		return nil, err
	}
	return index, nil
}

func DecodeCategories(data []byte) (*Categories, error) {
	categories := &Categories{}
	if err := decodeYAMLStrict(data, categories); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("categories.yaml: %s", err.Error())
	}
	if err := checkHeader("categories.yaml", categories.APIVersion, categories.Kind,
		KindTemplateCategories); err != nil {
		return nil, err
	}
	return categories, nil
}

func DecodeTags(data []byte) (*Tags, error) {
	tags := &Tags{}
	if err := decodeYAMLStrict(data, tags); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("tags.yaml: %s", err.Error())
	}
	if err := checkHeader("tags.yaml", tags.APIVersion, tags.Kind, KindTemplateTags); err != nil {
		return nil, err
	}
	return tags, nil
}

func checkHeader(file, apiVersion, kind, wantKind string) error {
	if apiVersion != APIVersion || kind != wantKind {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: expected apiVersion %q and kind %q", file, APIVersion, wantKind)
	}
	return nil
}
```

- [ ] **Step 5: Run the decoding tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/`
Expected: PASS for `TestDecodeTemplate`, `TestDecodeTemplateRefusesUnknownFields`, `TestDecodeIndex`, `TestDecodeIndexRefusesAnotherKindAndUnknownFields`, `TestDecodeCategoriesAndTags`, `TestIsCompatible`.

- [ ] **Step 6: Write the failing validation tests**

`hivepaas_app/service/apptemplateservice/templatemodel/validate_test.go`:

```go
package templatemodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func mustDecode(t *testing.T) *Template {
	t.Helper()
	tmpl, err := DecodeTemplate([]byte(validTemplateYAML))
	assert.NoError(t, err)
	return tmpl
}

func TestValidateAcceptsAValidTemplate(t *testing.T) {
	assert.NoError(t, mustDecode(t).Validate("demo"))
}

func TestValidateReportsEveryProblem(t *testing.T) {
	cases := map[string]struct {
		mutate func(tmpl *Template)
		want   string
	}{
		"wrong kind":            {func(tm *Template) { tm.Kind = "Spec" }, `kind must be "AppTemplate"`},
		"file name mismatch":    {func(tm *Template) { tm.Metadata.Name = "other" }, "must equal the file name"},
		"uppercase name":        {func(tm *Template) { tm.Metadata.Name = "Demo" }, "lowercase"},
		"long tagline":          {func(tm *Template) { tm.Metadata.Tagline = strings.Repeat("x", 81) }, "tagline"},
		"no categories":         {func(tm *Template) { tm.Metadata.Categories = nil }, "categories must have 1 to 3"},
		"flat category":         {func(tm *Template) { tm.Metadata.Categories = []string{"sql"} }, "parent/child"},
		"too many tags":         {func(tm *Template) { tm.Metadata.Tags = strings.Split("a,b,c,d,e,f,g,h,i", ",") }, "tags"},
		"jpeg icon":             {func(tm *Template) { tm.Metadata.Icon = "icons/demo.jpg" }, ".svg or .png"},
		"http link":             {func(tm *Template) { tm.Metadata.Links.Website = "http://example.com" }, "https"},
		"bad version code":      {func(tm *Template) { tm.Metadata.Requires.VersionCode = "1" }, "versionCode"},
		"secret default":        {func(tm *Template) { tm.Parameters[0].Default = "hunter2" }, "a secret has no default"},
		"short generate":        {func(tm *Template) { tm.Parameters[0].Generate.Length = 4 }, "generate.length"},
		"unknown charset":       {func(tm *Template) { tm.Parameters[0].Generate.Charset = "emoji" }, "charset"},
		"bad size bound":        {func(tm *Template) { tm.Parameters[1].Min = "lots" }, "min and max must be sizes"},
		"pattern on size":       {func(tm *Template) { tm.Parameters[1].Pattern = "^a$" }, "pattern does not apply"},
		"volume default":        {func(tm *Template) { tm.Parameters[2].Default = "vol" }, "a volume has no default"},
		"unknown type":          {func(tm *Template) { tm.Parameters[2].Type = "ssl-cert" }, "type \"ssl-cert\""},
		"duplicate parameter":   {func(tm *Template) { tm.Parameters[2].Name = "password" }, "declared twice"},
		"two default variants":  {func(tm *Template) { tm.Variants[1].Default = true }, "exactly one variant"},
		"two default versions":  {func(tm *Template) { tm.Versions[1].Default = true }, "exactly one version"},
		"deprecated default":    {func(tm *Template) { tm.Versions[0].Deprecated = true }, "cannot be deprecated"},
		"image with variants":   {func(tm *Template) { tm.Versions[0].Image = "demo:2.1.0" }, "set images and not image"},
		"undeclared variant":    {func(tm *Template) { tm.Versions[0].Images["musl"] = "demo:2.1.0-musl" }, `"musl" is not declared`},
		"floating tag":          {func(tm *Template) { tm.Versions[0].Images["alpine"] = "demo:2-alpine" }, "exact release"},
		"latest tag":            {func(tm *Template) { tm.Versions[0].Images["debian"] = "demo:latest" }, "exact release"},
		"empty override":        {func(tm *Template) { tm.Versions[0].Override = &Override{} }, "override.app is empty"},
		"no app":                {func(tm *Template) { tm.App = nil }, "app is required"},
		"no versions":           {func(tm *Template) { tm.Versions = nil }, "versions must not be empty"},
		"select without option": {func(tm *Template) {
			tm.Parameters = append(tm.Parameters, &Parameter{Name: "mode", Title: "Mode", Type: ParamTypeSelect})
		}, "options"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tmpl := mustDecode(t)
			tc.mutate(tmpl)
			err := tmpl.Validate("demo")
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
			assert.Contains(t, errorDetail(t, err), tc.want)
		})
	}
}

func TestValidateWithoutVariantsWantsImage(t *testing.T) {
	tmpl := mustDecode(t)
	tmpl.Variants = nil
	err := tmpl.Validate("demo")
	assert.Contains(t, errorDetail(t, err), "without variants, set image and not images")

	for _, version := range tmpl.Versions {
		version.Image, version.Images = version.Images["alpine"], nil
	}
	assert.NoError(t, tmpl.Validate("demo"))
}

func TestIsPinnedImage(t *testing.T) {
	for image, pinned := range map[string]bool{
		"postgres:17.6-alpine3.22":                    true,
		"postgres:17.6":                               true,
		"mariadb:11.8.3-noble":                        true,
		"minio/minio:RELEASE.2025-09-07T16-13-09Z":    true,
		"registry.example:5000/app:1.2.3":             true,
		"postgres@sha256:" + strings.Repeat("a", 64):  true,
		"postgres:17":                                 false,
		"postgres:17-alpine":                          false,
		"postgres:latest":                             false,
		"postgres":                                    false,
		"registry.example:5000/app":                   false,
	} {
		assert.Equal(t, pinned, IsPinnedImage(image), image)
	}
}

func TestParameterBounds(t *testing.T) {
	sizeParam := &Parameter{Type: ParamTypeSize, Min: "128MB", Max: "1GB"}
	minSize, maxSize := sizeParam.SizeBounds()
	assert.Equal(t, int64(128<<20), minSize.Bytes())
	assert.Equal(t, int64(1<<30), maxSize.Bytes())

	intParam := &Parameter{Type: ParamTypeInt, Min: 1, Max: float64(10)}
	minInt, maxInt := intParam.IntBounds()
	assert.Equal(t, int64(1), *minInt)
	assert.Equal(t, int64(10), *maxInt)

	unbounded := &Parameter{Type: ParamTypeInt}
	minInt, maxInt = unbounded.IntBounds()
	assert.Nil(t, minInt)
	assert.Nil(t, maxInt)
}
```

- [ ] **Step 7: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ -run 'Validate|IsPinnedImage|ParameterBounds'`
Expected: FAIL - `tmpl.Validate undefined`, `undefined: IsPinnedImage`.

- [ ] **Step 8: Write the validation**

`hivepaas_app/service/apptemplateservice/templatemodel/validate.go`:

```go
package templatemodel

import (
	"fmt"
	"maps"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	MaxTaglineLen = 80
	MaxCategories = 3
	MaxTags       = 8

	minGenerateLength = 8
	maxGenerateLength = 128
)

var (
	templateNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	paramNamePattern    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)
	choiceNamePattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)
	categoryRefPattern  = regexp.MustCompile(`^[a-z0-9-]+/[a-z0-9-]+$`)
	versionCodePattern  = regexp.MustCompile(`^v[0-9]{6}$`)
	// pinnedTagPattern is what separates 17.6-alpine from 17-alpine: a tag naming
	// one release carries at least major.minor, or a date the way minio's
	// RELEASE.2025-09-07T16-13-09Z does. It cannot know the patch level an image
	// publishes, so it is a floor, not proof.
	pinnedTagPattern = regexp.MustCompile(`[0-9]+\.[0-9]+|\.[0-9]{4}-[0-9]{2}-[0-9]{2}`)
)

// problems collects everything wrong with one file, so an author sees all of it
// at once instead of one problem per lint run.
type problems []string

func (p *problems) add(format string, args ...any) {
	*p = append(*p, fmt.Sprintf(format, args...))
}

func (p problems) err(name string) error {
	if len(p) == 0 {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s: %s", name, strings.Join(p, "; "))
}

// Validate checks what a template says about itself. What needs the rest of the
// repository - categories and tags against their vocabularies, the icon file -
// is templaterepo's, and whether the template renders is templaterender's.
//
// fileName is the file's name without .yaml, or empty when the template did not
// come from a file.
func (t *Template) Validate(fileName string) error {
	var p problems
	if t.APIVersion != APIVersion {
		p.add("apiVersion must be %q", APIVersion)
	}
	if t.Kind != KindAppTemplate {
		p.add("kind must be %q", KindAppTemplate)
	}
	t.Metadata.validate(fileName, &p)
	validateParameters(t.Parameters, &p)
	variants := validateVariants(t.Variants, &p)
	validateVersions(t.Versions, variants, &p)
	if len(t.App) == 0 {
		p.add("app is required")
	}

	name := t.Metadata.Name
	if name == "" {
		name = fileName
	}
	return p.err(name)
}

func (m *Metadata) validate(fileName string, p *problems) {
	switch {
	case !templateNamePattern.MatchString(m.Name):
		p.add("metadata.name %q must be lowercase letters, digits and dashes", m.Name)
	case fileName != "" && m.Name != fileName:
		p.add("metadata.name %q must equal the file name %q", m.Name, fileName)
	}
	if m.Title == "" {
		p.add("metadata.title is required")
	}
	if m.Tagline == "" || utf8.RuneCountInString(m.Tagline) > MaxTaglineLen {
		p.add("metadata.tagline is required and at most %d characters", MaxTaglineLen)
	}
	if strings.TrimSpace(m.Description) == "" {
		p.add("metadata.description is required")
	}
	if len(m.Categories) == 0 || len(m.Categories) > MaxCategories {
		p.add("metadata.categories must have 1 to %d entries", MaxCategories)
	}
	for _, category := range m.Categories {
		if !categoryRefPattern.MatchString(category) {
			p.add("metadata.categories: %q must be parent/child", category)
		}
	}
	if len(m.Tags) > MaxTags {
		p.add("metadata.tags must have at most %d entries", MaxTags)
	}
	if ext := path.Ext(m.Icon); ext != ".svg" && ext != ".png" {
		p.add("metadata.icon must be an .svg or .png file")
	}
	if m.Links != nil {
		for _, link := range [][2]string{
			{"website", m.Links.Website}, {"documentation", m.Links.Documentation}, {"source", m.Links.Source},
		} {
			if link[1] != "" && !isHTTPSURL(link[1]) {
				p.add("metadata.links.%s must be an https URL", link[0])
			}
		}
	}
	if !versionCodePattern.MatchString(m.Requires.VersionCode) {
		p.add("metadata.requires.versionCode must look like v000001")
	}
}

func isHTTPSURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func validateParameters(params []*Parameter, p *problems) {
	seen := map[string]bool{}
	for i, param := range params {
		if param == nil {
			p.add("parameters[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("parameters[%s]", param.Name)
		if !paramNamePattern.MatchString(param.Name) {
			p.add("parameters[%d].name %q must start with a letter and use letters, digits and _", i, param.Name)
		}
		if seen[param.Name] {
			p.add("%s is declared twice", prefix)
		}
		seen[param.Name] = true
		if param.Title == "" {
			p.add("%s.title is required", prefix)
		}
		validateParameterType(prefix, param, p)
	}
}

func validateParameterType(prefix string, param *Parameter, p *problems) {
	switch param.Type {
	case ParamTypeString:
		validateLengths(prefix, param, p)
		if param.Pattern != "" {
			if _, err := regexp.Compile(param.Pattern); err != nil {
				p.add("%s.pattern does not compile: %v", prefix, err)
			}
		}
		refuseAttributes(prefix, param, p, "min", "max", "generate", "options")
	case ParamTypeSecret:
		validateLengths(prefix, param, p)
		if param.Default != nil {
			p.add("%s: a secret has no default; use generate", prefix)
		}
		if gen := param.Generate; gen != nil {
			if gen.Length < minGenerateLength || gen.Length > maxGenerateLength {
				p.add("%s.generate.length must be %d to %d", prefix, minGenerateLength, maxGenerateLength)
			}
			if gen.Charset != "" && gen.Charset != CharsetAlnum && gen.Charset != CharsetHex {
				p.add("%s.generate.charset must be %s or %s", prefix, CharsetAlnum, CharsetHex)
			}
		}
		refuseAttributes(prefix, param, p, "pattern", "min", "max", "options")
	case ParamTypeInt:
		validateBounds(prefix, param, p, "integers", ToInt64)
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "generate", "options")
	case ParamTypeSize:
		validateBounds(prefix, param, p, "sizes", sizeBytes)
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "generate", "options")
	case ParamTypeBool:
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "options")
	case ParamTypeSelect:
		validateOptions(prefix, param.Options, p)
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate")
	case ParamTypeVolume:
		if param.Default != nil {
			p.add("%s: a volume has no default", prefix)
		}
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "options")
	default:
		p.add("%s.type %q is not one of %v", prefix, param.Type, AllParamTypes)
	}
}

func validateLengths(prefix string, param *Parameter, p *problems) {
	if param.MinLength != nil && *param.MinLength < 0 || param.MaxLength != nil && *param.MaxLength < 0 {
		p.add("%s.minLength and maxLength must not be negative", prefix)
	}
	if param.MinLength != nil && param.MaxLength != nil && *param.MinLength > *param.MaxLength {
		p.add("%s.minLength must not exceed maxLength", prefix)
	}
}

func validateBounds(prefix string, param *Parameter, p *problems, kind string, parse func(any) (int64, bool)) {
	minValue, minOK := parse(param.Min)
	maxValue, maxOK := parse(param.Max)
	if param.Min != nil && !minOK || param.Max != nil && !maxOK {
		p.add("%s.min and max must be %s", prefix, kind)
		return
	}
	if param.Min != nil && param.Max != nil && minValue > maxValue {
		p.add("%s.min must not exceed max", prefix)
	}
}

func validateOptions(prefix string, options []*SelectOption, p *problems) {
	if len(options) == 0 {
		p.add("%s.options must list at least one option", prefix)
		return
	}
	seen := map[string]bool{}
	for _, option := range options {
		if option == nil || option.Value == "" || option.Title == "" {
			p.add("%s.options: every option needs a value and a title", prefix)
			continue
		}
		if seen[option.Value] {
			p.add("%s.options: %q is listed twice", prefix, option.Value)
		}
		seen[option.Value] = true
	}
}

func refuseAttributes(prefix string, param *Parameter, p *problems, names ...string) {
	set := map[string]bool{
		"pattern":   param.Pattern != "",
		"minLength": param.MinLength != nil,
		"maxLength": param.MaxLength != nil,
		"min":       param.Min != nil,
		"max":       param.Max != nil,
		"generate":  param.Generate != nil,
		"options":   len(param.Options) > 0,
	}
	for _, name := range names {
		if set[name] {
			p.add("%s.%s does not apply to type %s", prefix, name, param.Type)
		}
	}
}

func validateVariants(variants []*Variant, p *problems) map[string]bool {
	names := map[string]bool{}
	defaults := 0
	for i, variant := range variants {
		if variant == nil {
			p.add("variants[%d] is empty", i)
			continue
		}
		if !choiceNamePattern.MatchString(variant.Name) {
			p.add("variants[%d].name %q is invalid", i, variant.Name)
		}
		if names[variant.Name] {
			p.add("variant %q is declared twice", variant.Name)
		}
		names[variant.Name] = true
		if variant.Title == "" {
			p.add("variant %q needs a title", variant.Name)
		}
		if variant.Default {
			defaults++
		}
	}
	if len(variants) > 0 && defaults != 1 {
		p.add("exactly one variant must be the default")
	}
	return names
}

func validateVersions(versions []*Version, variants map[string]bool, p *problems) {
	if len(versions) == 0 {
		p.add("versions must not be empty")
		return
	}
	seen := map[string]bool{}
	defaults := 0
	for i, version := range versions {
		if version == nil {
			p.add("versions[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("versions[%s]", version.Name)
		if !choiceNamePattern.MatchString(version.Name) {
			p.add("versions[%d].name %q is invalid", i, version.Name)
		}
		if seen[version.Name] {
			p.add("%s is declared twice", prefix)
		}
		seen[version.Name] = true
		if version.Release == "" {
			p.add("%s.release is required", prefix)
		}
		if version.Default {
			defaults++
			if version.Deprecated {
				p.add("%s: the default version cannot be deprecated", prefix)
			}
		}
		validateVersionImages(prefix, version, variants, p)
		if version.Override != nil && len(version.Override.App) == 0 {
			p.add("%s.override.app is empty", prefix)
		}
	}
	if defaults != 1 {
		p.add("exactly one version must be the default")
	}
}

func validateVersionImages(prefix string, version *Version, variants map[string]bool, p *problems) {
	if len(variants) == 0 {
		if version.Image == "" || len(version.Images) > 0 {
			p.add("%s: without variants, set image and not images", prefix)
		}
		checkPinned(prefix+".image", version.Image, p)
		return
	}
	if version.Image != "" || len(version.Images) == 0 {
		p.add("%s: with variants, set images and not image", prefix)
	}
	for _, variant := range slices.Sorted(maps.Keys(version.Images)) {
		if !variants[variant] {
			p.add("%s.images: variant %q is not declared", prefix, variant)
		}
		checkPinned(fmt.Sprintf("%s.images[%s]", prefix, variant), version.Images[variant], p)
	}
}

func checkPinned(field, image string, p *problems) {
	if image != "" && !IsPinnedImage(image) {
		p.add("%s %q must name an exact release, not a moving tag", field, image)
	}
}

// IsPinnedImage reports whether an image reference names one release: a digest,
// or a tag carrying at least major.minor. A moving tag changes what runs without
// any template revision changing - on a redeploy, or when swarm reschedules a
// task onto another node.
func IsPinnedImage(image string) bool {
	ref := imageref.Parse(image)
	if ref.Digest != "" {
		return true
	}
	if ref.Tag == "" || ref.Tag == "latest" {
		return false
	}
	return pinnedTagPattern.MatchString(ref.Tag)
}

// ToInt64 accepts the integer shapes a YAML or JSON decode produces.
func ToInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case uint64:
		if n > uint64(1<<63-1) {
			return 0, false
		}
		return int64(n), true
	case float64:
		if n != float64(int64(n)) {
			return 0, false
		}
		return int64(n), true
	default:
		return 0, false
	}
}

func sizeBytes(v any) (int64, bool) {
	text, ok := v.(string)
	if !ok {
		return 0, false
	}
	size, err := unit.ParseDataSizeString(text)
	if err != nil {
		return 0, false
	}
	return size.Bytes(), true
}

// IntBounds returns an int parameter's min and max, nil where unset. It assumes
// Validate passed.
func (p *Parameter) IntBounds() (minValue, maxValue *int64) {
	if v, ok := ToInt64(p.Min); ok {
		minValue = &v
	}
	if v, ok := ToInt64(p.Max); ok {
		maxValue = &v
	}
	return minValue, maxValue
}

// SizeBounds returns a size parameter's min and max, nil where unset. It assumes
// Validate passed.
func (p *Parameter) SizeBounds() (minValue, maxValue *unit.DataSize) {
	if v, ok := sizeBytes(p.Min); ok {
		size := unit.DataSize(v)
		minValue = &size
	}
	if v, ok := sizeBytes(p.Max); ok {
		size := unit.DataSize(v)
		maxValue = &size
	}
	return minValue, maxValue
}
```

- [ ] **Step 9: Run the package tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/`
Expected: PASS.

- [ ] **Step 10: Lint and commit**

Run: `go build ./... && make lint-local`
Expected: `0 issues.` (errcodelint finds `ERR_APP_TEMPLATE_INVALID` declared, referenced and translated).

```bash
git add hivepaas_app/service/apptemplateservice/templatemodel hivepaas_app/hperrors/errors_app_template.go \
  hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml
git commit -m "feat(apptemplate): template format types, strict decoding and validation"
```

---
## Task 2: What phase 1 can build - `specmodel.CheckBuildable`

**Files:**
- Create: `hivepaas_app/service/specservice/specmodel/buildable.go`
- Modify: `hivepaas_app/hperrors/errors_spec.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.spec.en.toml`
- Test: `hivepaas_app/service/specservice/specmodel/buildable_test.go`

**Interfaces:**
- Consumes: `specmodel.AppDoc`, `specmodel.SingletonBlockName`.
- Produces:
  - `type Block string` with `BlockDeploymentSource`, `BlockDeploymentStorage`, `BlockContainerHealthcheck`, `BlockDeploymentResources`, `BlockSettingsKind`, `BlockSettingsEnvVars`, `BlockSettingsRouting`; `var BuildableBlocks []Block`
  - `func CheckBuildable(doc *AppDoc) error` - `ErrSpecBlockUnsupported`, the offending path in the extra detail
  - `func PresentBlocks(doc *AppDoc) []Block` - the buildable blocks a document carries, in `BuildableBlocks` order
  - `hperrors.ErrSpecBlockUnsupported`

The check lives in `specmodel` rather than beside the builders (Task 9) because rendering (Task 5) and the linter (Task 14) must refuse the same things the builders cannot build, and neither may depend on a service.

- [ ] **Step 1: Declare the error**

Append to the `var (...)` block in `hivepaas_app/hperrors/errors_spec.go`:

```go
	ErrSpecBlockUnsupported        = NewErr(ErrUnsupported, "ERR_SPEC_BLOCK_UNSUPPORTED")
```

Append to `hivepaas_app/pkg/translation/messages/en/errors.spec.en.toml`:

```toml
ERR_SPEC_BLOCK_UNSUPPORTED = "Part of this configuration cannot be built into an app yet"
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/specservice/specmodel/buildable_test.go`:

```go
package specmodel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// buildableDocYAML uses every block and field phase 1 can build.
const buildableDocYAML = `
deployment:
  source:
    activeMethod: image
    imageSource: {image: "postgres:17.6-alpine3.22"}
    command: postgres -c max_connections=200
    workingDir: /
  storage:
    mounts:
      /var/lib/postgresql/data: {type: volume, source: vol-1, readOnly: false, volumeOptions: {subpath: data}}
  container:
    healthcheck: {enabled: true, mode: CMD-SHELL, command: pg_isready, interval: 10s, retries: 5}
  resources:
    reservations: {cpus: 0.5, memory: 256mb}
    limits: {cpus: 1, memory: 512mb, pids: 100}
settings:
  kind: {category: database, engine: postgres}
  envVars: {data: [{k: A, v: b}]}
  routing: {port: 5432}
`

func decodeDoc(t *testing.T, text string) *AppDoc {
	t.Helper()
	doc := &AppDoc{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), doc))
	return doc
}

func buildableErrorDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestCheckBuildableAcceptsTheSupportedSubset(t *testing.T) {
	assert.NoError(t, CheckBuildable(decodeDoc(t, buildableDocYAML)))
	assert.NoError(t, CheckBuildable(&AppDoc{}))
	assert.NoError(t, CheckBuildable(nil))
}

func TestPresentBlocks(t *testing.T) {
	assert.Equal(t, BuildableBlocks, PresentBlocks(decodeDoc(t, buildableDocYAML)))
	assert.Equal(t, []Block{BlockSettingsKind}, PresentBlocks(decodeDoc(t, "settings:\n  kind: {category: cache}\n")))
	assert.Empty(t, PresentBlocks(&AppDoc{}))
}

func TestCheckBuildableRefusesTheRest(t *testing.T) {
	cases := map[string]struct {
		doc  string
		path string
	}{
		"app name":           {"name: db\n", "name"},
		"repository build":   {"deployment:\n  source:\n    activeMethod: repo\n", "deployment.source.activeMethod"},
		"registry auth":      {"deployment:\n  source:\n    imageSource: {image: x, registryAuth: {id: r}}\n", "deployment.source.imageSource.registryAuth"},
		"unknown source key": {"deployment:\n  source:\n    preDeploymentCommand: x\n", "deployment.source.preDeploymentCommand"},
		"container user":     {"deployment:\n  container:\n    user: root\n", "deployment.container.user"},
		"memory swap":        {"deployment:\n  resources:\n    memory: {swap: 1gb}\n", "deployment.resources.memory"},
		"generic resources":  {"deployment:\n  resources:\n    reservations: {genericResources: [{kind: gpu, value: '1'}]}\n", "deployment.resources.reservations.genericResources"},
		"bind mount":         {"deployment:\n  storage:\n    mounts:\n      /data: {type: bind, source: /srv}\n", "deployment.storage.mounts./data.type"},
		"volume labels":      {"deployment:\n  storage:\n    mounts:\n      /data: {type: volume, source: v, volumeOptions: {labels: {a: b}}}\n", "deployment.storage.mounts./data.volumeOptions.labels"},
		"networks":           {"deployment:\n  networks:\n    dnsConfig: {nameservers: [1.1.1.1]}\n", "deployment.networks"},
		"service mode":       {"deployment:\n  service:\n    modeSpec: {mode: global}\n", "deployment.service"},
		"config files":       {"settings:\n  configFiles: {}\n", "settings.configFiles"},
		"routing domains":    {"settings:\n  routing: {port: 80, domains: []}\n", "settings.routing.domains"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckBuildable(decodeDoc(t, tc.doc))
			assert.ErrorIs(t, err, hperrors.ErrSpecBlockUnsupported)
			assert.Contains(t, buildableErrorDetail(t, err), tc.path)
		})
	}
}
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -run 'Buildable|PresentBlocks'`
Expected: FAIL - `undefined: CheckBuildable`.

- [ ] **Step 4: Write the check**

`hivepaas_app/service/specservice/specmodel/buildable.go`:

```go
package specmodel

import (
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// Block names a part of an AppDoc that can be built into an app.
type Block string

const (
	BlockDeploymentSource     Block = "deployment.source"
	BlockDeploymentStorage    Block = "deployment.storage"
	BlockContainerHealthcheck Block = "deployment.container.healthcheck"
	BlockDeploymentResources  Block = "deployment.resources"
	BlockSettingsKind         Block = "settings.kind"
	BlockSettingsEnvVars      Block = "settings.envVars"
	BlockSettingsRouting      Block = "settings.routing"
)

// BuildableBlocks is every block specservice.BuildApp builds, in the order it
// builds them. specserviceimpl's builder registry is tested against this list.
//
// TODO: app templates phase 3 - more blocks: repository builds, registry auth,
// bind mounts, config files, secrets, scheduled jobs, routing domains. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
var BuildableBlocks = []Block{BlockDeploymentSource, BlockDeploymentStorage, BlockContainerHealthcheck,
	BlockDeploymentResources, BlockSettingsKind, BlockSettingsEnvVars, BlockSettingsRouting}

// CheckBuildable refuses any part of doc that phase 1 cannot build.
//
// Refusing is the point: a field that was skipped instead would provision an
// app with less than the document describes, and nothing would say which part
// went missing. The error's extra detail names the first offending path.
func CheckBuildable(doc *AppDoc) error {
	if doc == nil {
		return nil
	}
	if err := onlyFields("", doc, "deployment", "settings"); err != nil {
		return err
	}
	if err := checkDeployment(doc.Deployment); err != nil {
		return err
	}
	return checkSettings(doc.Settings)
}

// PresentBlocks lists the buildable blocks doc carries, in BuildableBlocks order.
// It assumes CheckBuildable passed.
func PresentBlocks(doc *AppDoc) []Block {
	var blocks []Block
	if doc == nil {
		return blocks
	}
	if d := doc.Deployment; d != nil {
		if d.Source != nil {
			blocks = append(blocks, BlockDeploymentSource)
		}
		if d.Storage != nil && len(d.Storage.Mounts) > 0 {
			blocks = append(blocks, BlockDeploymentStorage)
		}
		if d.Container != nil && d.Container.Healthcheck != nil {
			blocks = append(blocks, BlockContainerHealthcheck)
		}
		if d.Resources != nil {
			blocks = append(blocks, BlockDeploymentResources)
		}
	}
	for _, pair := range []struct {
		typ   base.SettingType
		block Block
	}{
		{base.SettingTypeAppKind, BlockSettingsKind},
		{base.SettingTypeEnvVar, BlockSettingsEnvVars},
		{base.SettingTypeAppRouting, BlockSettingsRouting},
	} {
		if _, ok := doc.Settings[SingletonBlockName(pair.typ)]; ok {
			blocks = append(blocks, pair.block)
		}
	}
	return blocks
}

func checkDeployment(d *Deployment) error {
	if d == nil {
		return nil
	}
	if err := onlyFields("deployment.", d, "source", "container", "resources", "storage"); err != nil {
		return err
	}
	if err := checkSource(d.Source); err != nil {
		return err
	}
	if d.Container != nil {
		if err := onlyFields("deployment.container.", d.Container, "healthcheck"); err != nil {
			return err
		}
	}
	if err := checkResources(d.Resources); err != nil {
		return err
	}
	return checkStorage(d.Storage)
}

func checkSource(source map[string]any) error {
	for _, key := range slices.Sorted(maps.Keys(source)) {
		path := "deployment.source." + key
		switch key {
		case "activeMethod":
			if source[key] != string(base.DeploymentMethodImage) {
				return unsupported(path)
			}
		case "imageSource":
			imageSource, ok := source[key].(map[string]any)
			if !ok {
				return unsupported(path)
			}
			for _, field := range slices.Sorted(maps.Keys(imageSource)) {
				if field != "image" {
					return unsupported(path + "." + field)
				}
			}
		case "command", "workingDir":
		default:
			return unsupported(path)
		}
	}
	return nil
}

func checkResources(r *Resources) error {
	if r == nil {
		return nil
	}
	if err := onlyFields("deployment.resources.", r, "reservations", "limits"); err != nil {
		return err
	}
	if r.Reservations != nil {
		if err := onlyFields("deployment.resources.reservations.", r.Reservations, "cpus", "memory"); err != nil {
			return err
		}
	}
	if r.Limits != nil {
		return onlyFields("deployment.resources.limits.", r.Limits, "cpus", "memory", "pids")
	}
	return nil
}

func checkStorage(s *Storage) error {
	if s == nil {
		return nil
	}
	if err := onlyFields("deployment.storage.", s, "mounts"); err != nil {
		return err
	}
	for _, target := range slices.Sorted(maps.Keys(s.Mounts)) {
		m := s.Mounts[target]
		path := "deployment.storage.mounts." + target + "."
		if m.Type != mount.TypeVolume {
			return unsupported(path + "type")
		}
		if err := onlyFields(path, &m, "type", "source", "readOnly", "volumeOptions"); err != nil {
			return err
		}
		if m.VolumeOptions != nil {
			if err := onlyFields(path+"volumeOptions.", m.VolumeOptions, "subpath"); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkSettings(settings map[string]any) error {
	kind := SingletonBlockName(base.SettingTypeAppKind)
	envVars := SingletonBlockName(base.SettingTypeEnvVar)
	routing := SingletonBlockName(base.SettingTypeAppRouting)

	for _, key := range slices.Sorted(maps.Keys(settings)) {
		switch key {
		case kind, envVars:
		case routing:
			body, ok := settings[key].(map[string]any)
			if !ok {
				return unsupported("settings." + key)
			}
			for _, field := range slices.Sorted(maps.Keys(body)) {
				if field != "port" {
					return unsupported("settings." + key + "." + field)
				}
			}
		default:
			return unsupported("settings." + key)
		}
	}
	return nil
}

// onlyFields refuses the first non-zero field of a struct not named in allowed,
// naming it by its yaml key under prefix.
func onlyFields(prefix string, value any, allowed ...string) error {
	v := reflect.Indirect(reflect.ValueOf(value))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil
	}
	typ := v.Type()
	for i := range typ.NumField() {
		field := typ.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		if name == "" {
			name = field.Name
		}
		if slices.Contains(allowed, name) || v.Field(i).IsZero() {
			continue
		}
		return unsupported(prefix + name)
	}
	return nil
}

func unsupported(path string) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail("%s", path)
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/specservice/specmodel/`
Expected: PASS, including the existing singleton, key and manifest tests.

- [ ] **Step 6: Lint and commit**

Run: `go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/specservice/specmodel/buildable.go hivepaas_app/service/specservice/specmodel/buildable_test.go \
  hivepaas_app/hperrors/errors_spec.go hivepaas_app/pkg/translation/messages/en/errors.spec.en.toml
git commit -m "feat(spec): declare the AppDoc subset that can be built into an app"
```

---

## Task 3: Merge patch and placeholders - `templaterender`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templaterender/mergepatch.go`
- Create: `hivepaas_app/service/apptemplateservice/templaterender/placeholder.go`
- Test: `hivepaas_app/service/apptemplateservice/templaterender/mergepatch_test.go`
- Test: `hivepaas_app/service/apptemplateservice/templaterender/placeholder_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces:
  - `func MergePatch(target, patch any) any` - RFC 7396; neither argument is modified
  - `type Resolver func(ref string) (value any, keep bool, err error)`
  - `func Substitute(node any, resolve Resolver) (any, error)`
  - `func deepCopy(node any) any` (package-private, used by Task 5)

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/service/apptemplateservice/templaterender/mergepatch_test.go`:

```go
package templaterender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergePatch(t *testing.T) {
	target := map[string]any{
		"a":    1,
		"b":    map[string]any{"c": 2, "d": 3},
		"list": []any{1, 2},
	}
	patch := map[string]any{
		"b":    map[string]any{"c": nil, "e": 4},
		"list": []any{9},
		"f":    "new",
	}

	got := MergePatch(target, patch)

	assert.Equal(t, map[string]any{
		"a":    1,
		"b":    map[string]any{"d": 3, "e": 4},
		"list": []any{9},
		"f":    "new",
	}, got, "maps merge, null deletes, lists replace")
	assert.Equal(t, map[string]any{"c": 2, "d": 3}, target["b"], "the target is not modified")
}

func TestMergePatchNonMaps(t *testing.T) {
	assert.Equal(t, map[string]any{"a": 1}, MergePatch("scalar", map[string]any{"a": 1}))
	assert.Equal(t, "scalar", MergePatch(map[string]any{"a": 1}, "scalar"))
}

func TestMergePatchCopiesWhatItKeeps(t *testing.T) {
	inner := map[string]any{"x": 1}
	got, ok := MergePatch(map[string]any{"inner": inner}, map[string]any{"other": 2}).(map[string]any)
	assert.True(t, ok)
	got["inner"].(map[string]any)["x"] = 99
	assert.Equal(t, 1, inner["x"], "the result shares nothing with its inputs")
}
```

`hivepaas_app/service/apptemplateservice/templaterender/placeholder_test.go`:

```go
package templaterender

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

var errTestUndefined = errors.New("undefined")

func mapResolver(values map[string]any) Resolver {
	return func(ref string) (any, bool, error) {
		value, ok := values[ref]
		if !ok {
			return nil, false, errTestUndefined
		}
		return value, false, nil
	}
}

func TestSubstitute(t *testing.T) {
	resolve := mapResolver(map[string]any{
		"params.memory":  unit.DataSize(512 << 20),
		"params.user":    "app",
		"params.port":    int64(5432),
		"params.enabled": true,
		"version.name":   "17",
	})
	tree := map[string]any{
		"memory":  "${{ params.memory }}",
		"port":    "${{params.port}}",
		"enabled": "${{ params.enabled }}",
		"cmd":     "pg_isready -U ${{ params.user }} -p ${{ params.port }}",
		"nested":  []any{map[string]any{"v": "${{ version.name }}"}},
		"escaped": "echo $${{ params.user }}",
		"envRef":  "${HIVEPAAS_PASSWORD}",
		"number":  3,
	}

	got, err := Substitute(tree, resolve)

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"memory":  unit.DataSize(512 << 20), // a whole-string placeholder takes the value's type
		"port":    int64(5432),
		"enabled": true,
		"cmd":     "pg_isready -U app -p 5432",
		"nested":  []any{map[string]any{"v": "17"}},
		"escaped": "echo ${{ params.user }}",
		"envRef":  "${HIVEPAAS_PASSWORD}", // HivePaaS's own references are not placeholders
		"number":  3,
	}, got)
	assert.Equal(t, "${{ params.memory }}", tree["memory"], "the input tree is not modified")
}

func TestSubstituteKeepsWhatTheResolverKeeps(t *testing.T) {
	resolve := func(ref string) (any, bool, error) {
		if ref == "params.password" {
			return nil, true, nil
		}
		return "app", false, nil
	}
	got, err := Substitute(map[string]any{
		"whole":    "${{params.password}}",
		"embedded": "postgres://${{ params.user }}:${{  params.password }}@db",
	}, resolve)

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"whole":    "${{ params.password }}",
		"embedded": "postgres://app:${{ params.password }}@db",
	}, got, "a kept placeholder is written in one canonical spelling")
}

func TestSubstituteReportsTheResolverError(t *testing.T) {
	_, err := Substitute([]any{"ok", "${{ params.nope }}"}, mapResolver(map[string]any{}))
	assert.ErrorIs(t, err, errTestUndefined)
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/`
Expected: FAIL - `undefined: MergePatch`, `undefined: Substitute`.

- [ ] **Step 3: Write the merge patch**

`hivepaas_app/service/apptemplateservice/templaterender/mergepatch.go`:

```go
// Package templaterender turns a template and a user's choices into a
// specmodel.AppDoc. It is pure - no database, no network - so the linter in
// tools/apptemplate renders templates exactly the way HivePaaS does.
package templaterender

// MergePatch applies patch to target as an RFC 7396 JSON Merge Patch: maps
// merge recursively, a nil value deletes its key, and anything else - lists
// included - replaces wholesale. It returns a new tree and modifies neither
// argument.
func MergePatch(target, patch any) any {
	patchMap, isMap := patch.(map[string]any)
	if !isMap {
		return deepCopy(patch)
	}

	out := map[string]any{}
	if targetMap, ok := target.(map[string]any); ok {
		for key, value := range targetMap {
			out[key] = deepCopy(value)
		}
	}
	for key, value := range patchMap {
		if value == nil {
			delete(out, key)
			continue
		}
		out[key] = MergePatch(out[key], value)
	}
	return out
}

// deepCopy copies the maps and lists a decoded YAML tree is made of. Scalars
// are values already.
func deepCopy(node any) any {
	switch v := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			out[key] = deepCopy(value)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, value := range v {
			out[i] = deepCopy(value)
		}
		return out
	default:
		return v
	}
}
```

- [ ] **Step 4: Write the placeholders**

`hivepaas_app/service/apptemplateservice/templaterender/placeholder.go`:

```go
package templaterender

import (
	"fmt"
	"regexp"
	"strings"
)

// placeholderPattern matches ${{ ref }}, and $${{ ref }} as its escape. ${VAR}
// is not a placeholder: it is HivePaaS's own environment variable reference,
// and it must reach the app untouched.
var placeholderPattern = regexp.MustCompile(`\$?\$\{\{\s*([A-Za-z0-9_.]+)\s*\}\}`)

// Resolver answers one placeholder. keep asks for the placeholder to stay in the
// output as written - how a base render leaves secrets out.
type Resolver func(ref string) (value any, keep bool, err error)

// Substitute replaces placeholders in every string of a tree, returning a new
// tree. A string that is exactly one placeholder takes the resolved value's own
// type, so a size stays a size and a number stays a number; a placeholder inside
// a longer string is formatted into it.
func Substitute(node any, resolve Resolver) (any, error) {
	switch v := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			substituted, err := Substitute(value, resolve)
			if err != nil {
				return nil, err
			}
			out[key] = substituted
		}
		return out, nil
	case []any:
		out := make([]any, len(v))
		for i, value := range v {
			substituted, err := Substitute(value, resolve)
			if err != nil {
				return nil, err
			}
			out[i] = substituted
		}
		return out, nil
	case string:
		return substituteString(v, resolve)
	default:
		return v, nil
	}
}

func substituteString(text string, resolve Resolver) (any, error) {
	matches := placeholderPattern.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}

	whole := matches[0]
	if len(matches) == 1 && whole[0] == 0 && whole[1] == len(text) && !isEscaped(text, whole) {
		ref := text[whole[2]:whole[3]]
		value, keep, err := resolve(ref)
		if err != nil {
			return nil, err
		}
		if keep {
			return canonicalPlaceholder(ref), nil
		}
		return value, nil
	}

	var out strings.Builder
	last := 0
	for _, match := range matches {
		out.WriteString(text[last:match[0]])
		last = match[1]
		ref := text[match[2]:match[3]]
		if isEscaped(text, match) {
			out.WriteString(text[match[0]+1 : match[1]])
			continue
		}
		value, keep, err := resolve(ref)
		if err != nil {
			return nil, err
		}
		if keep {
			out.WriteString(canonicalPlaceholder(ref))
			continue
		}
		out.WriteString(fmt.Sprint(value))
	}
	out.WriteString(text[last:])
	return out.String(), nil
}

func isEscaped(text string, match []int) bool {
	return strings.HasPrefix(text[match[0]:], "$$")
}

func canonicalPlaceholder(ref string) string {
	return "${{ " + ref + " }}"
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/`
Expected: PASS.

- [ ] **Step 6: Lint and commit**

Run: `go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templaterender
git commit -m "feat(apptemplate): JSON merge patch and typed placeholder substitution"
```

---
## Task 4: Parameters - `templaterender.ResolveParams`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templaterender/params.go`
- Modify: `hivepaas_app/hperrors/errors_app_template.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`
- Test: `hivepaas_app/service/apptemplateservice/templaterender/params_test.go`

**Interfaces:**
- Consumes: `templatemodel.Parameter`, `templatemodel.ParamType*`, `templatemodel.ToInt64`, `(*Parameter).IntBounds`, `(*Parameter).SizeBounds`, `templatemodel.CharsetHex` (Task 1).
- Produces:
  - `type Value struct { Param *templatemodel.Parameter; Value any; Generated bool }` - `Value` is `string`, `int64`, `bool` or `unit.DataSize` by type, nil for an optional parameter nobody set
  - `func (v *Value) Text() string`
  - `func ResolveParams(defs []*templatemodel.Parameter, input map[string]any) (map[string]*Value, error)`
  - `func ValidateDefaults(defs []*templatemodel.Parameter) error`
  - `hperrors.ErrAppTemplateParamInvalid`

Input values arrive from JSON (a number is a `float64`) and from the CLI (everything is a string), so an `int` or `bool` accepts both its own type and its text.

- [ ] **Step 1: Declare the error**

Add to the `var (...)` block in `hivepaas_app/hperrors/errors_app_template.go`:

```go
	ErrAppTemplateParamInvalid = NewErr(ErrArgumentInvalid, "ERR_APP_TEMPLATE_PARAM_INVALID")
```

Add to `errors.app_template.en.toml`:

```toml
ERR_APP_TEMPLATE_PARAM_INVALID = "Parameter '{{.Name}}' is invalid"
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/apptemplateservice/templaterender/params_test.go`:

```go
package templaterender

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

func intPtr(v int) *int { return &v }

func testParams() []*templatemodel.Parameter {
	return []*templatemodel.Parameter{
		{Name: "dbName", Type: templatemodel.ParamTypeString, Default: "app", Pattern: "^[a-z_]+$",
			MaxLength: intPtr(8)},
		{Name: "password", Type: templatemodel.ParamTypeSecret,
			Generate: &templatemodel.Generate{Length: 24}},
		{Name: "token", Type: templatemodel.ParamTypeSecret, Optional: true},
		{Name: "workers", Type: templatemodel.ParamTypeInt, Default: 4, Min: 1, Max: 64},
		{Name: "memoryLimit", Type: templatemodel.ParamTypeSize, Default: "512MB", Min: "128MB"},
		{Name: "debug", Type: templatemodel.ParamTypeBool, Default: false},
		{Name: "mode", Type: templatemodel.ParamTypeSelect, Default: "fast",
			Options: []*templatemodel.SelectOption{{Value: "fast", Title: "Fast"}, {Value: "safe", Title: "Safe"}}},
		{Name: "dataVolume", Type: templatemodel.ParamTypeVolume},
	}
}

func paramDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestResolveParamsFillsDefaultsAndGeneratesSecrets(t *testing.T) {
	values, err := ResolveParams(testParams(), map[string]any{"dataVolume": "vol-1"})

	assert.NoError(t, err)
	assert.Equal(t, "app", values["dbName"].Value)
	assert.Equal(t, int64(4), values["workers"].Value)
	assert.Equal(t, unit.DataSize(512<<20), values["memoryLimit"].Value)
	assert.Equal(t, false, values["debug"].Value)
	assert.Equal(t, "fast", values["mode"].Value)
	assert.Equal(t, "vol-1", values["dataVolume"].Value)
	assert.Nil(t, values["token"].Value, "an optional parameter nobody set stays unset")
	assert.Equal(t, "", values["token"].Text())

	password := values["password"]
	assert.True(t, password.Generated)
	assert.Regexp(t, regexp.MustCompile(`^[A-Za-z0-9]{24}$`), password.Text())

	again, err := ResolveParams(testParams(), map[string]any{"dataVolume": "vol-1"})
	assert.NoError(t, err)
	assert.NotEqual(t, password.Text(), again["password"].Text(), "every generated secret is new")
}

func TestResolveParamsConvertsInput(t *testing.T) {
	values, err := ResolveParams(testParams(), map[string]any{
		"dbName":      "shop",
		"password":    "given-password",
		"workers":     float64(16), // what a JSON body decodes to
		"memoryLimit": "1GB",
		"debug":       "true", // what the CLI passes
		"mode":        "safe",
		"dataVolume":  "vol-2",
	})

	assert.NoError(t, err)
	assert.Equal(t, "shop", values["dbName"].Value)
	assert.Equal(t, "given-password", values["password"].Value)
	assert.False(t, values["password"].Generated)
	assert.Equal(t, int64(16), values["workers"].Value)
	assert.Equal(t, "16", values["workers"].Text())
	assert.Equal(t, unit.DataSize(1<<30), values["memoryLimit"].Value)
	assert.Equal(t, "1gb", values["memoryLimit"].Text())
	assert.Equal(t, true, values["debug"].Value)
}

func TestResolveParamsRefuses(t *testing.T) {
	cases := map[string]struct {
		input map[string]any
		want  string
	}{
		"unknown parameter": {map[string]any{"dataVolume": "v", "surprise": 1}, "surprise"},
		"missing volume":    {map[string]any{}, "dataVolume: a value is required"},
		"pattern":           {map[string]any{"dataVolume": "v", "dbName": "Shop"}, "dbName: must match"},
		"too long":          {map[string]any{"dataVolume": "v", "dbName": "abcdefghi"}, "dbName: must be at most 8"},
		"not a number":      {map[string]any{"dataVolume": "v", "workers": "many"}, "workers: must be a whole number"},
		"fraction":          {map[string]any{"dataVolume": "v", "workers": 1.5}, "workers: must be a whole number"},
		"above max":         {map[string]any{"dataVolume": "v", "workers": 65}, "workers: must be at most 64"},
		"below min size":    {map[string]any{"dataVolume": "v", "memoryLimit": "64MB"}, "memoryLimit: must be at least"},
		"not a size":        {map[string]any{"dataVolume": "v", "memoryLimit": "lots"}, "memoryLimit: must be a size"},
		"not a bool":        {map[string]any{"dataVolume": "v", "debug": "maybe"}, "debug: must be true or false"},
		"unknown option":    {map[string]any{"dataVolume": "v", "mode": "turbo"}, "mode: must be one of"},
		"secret not text":   {map[string]any{"dataVolume": "v", "password": 42}, "password: must be text"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ResolveParams(testParams(), tc.input)
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
			assert.Contains(t, paramDetail(t, err), tc.want)
		})
	}
}

func TestResolveParamsGeneratesHex(t *testing.T) {
	defs := []*templatemodel.Parameter{{Name: "key", Type: templatemodel.ParamTypeSecret,
		Generate: &templatemodel.Generate{Length: 32, Charset: templatemodel.CharsetHex}}}
	values, err := ResolveParams(defs, nil)
	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), values["key"].Text())
}

func TestValidateDefaults(t *testing.T) {
	assert.NoError(t, ValidateDefaults(testParams()))

	broken := testParams()
	broken[4].Default = "64MB" // below its own min
	err := ValidateDefaults(broken)
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
	assert.Contains(t, paramDetail(t, err), "memoryLimit")
}
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/ -run 'ResolveParams|ValidateDefaults'`
Expected: FAIL - `undefined: ResolveParams`.

- [ ] **Step 4: Write the parameters**

`hivepaas_app/service/apptemplateservice/templaterender/params.go`:

```go
package templaterender

import (
	"crypto/rand"
	"fmt"
	"maps"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	alnumAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	hexAlphabet   = "0123456789abcdef"
)

// Value is one parameter resolved to its canonical, typed value.
type Value struct {
	Param *templatemodel.Parameter
	// Value is a string, int64, bool or unit.DataSize according to Param.Type,
	// or nil for an optional parameter nobody set.
	Value     any
	Generated bool
}

// Text is the value as text: what the app-template setting stores, and what a
// placeholder inside a longer string is replaced with.
func (v *Value) Text() string {
	if v.Value == nil {
		return ""
	}
	return fmt.Sprint(v.Value)
}

// ResolveParams checks input against a template's parameters, fills in defaults
// and generates the secrets nobody gave. A parameter the template does not
// declare is refused rather than ignored: it is a form and a template that
// disagree, and ignoring it would provision without the value somebody meant.
func ResolveParams(defs []*templatemodel.Parameter, input map[string]any) (map[string]*Value, error) {
	declared := make(map[string]bool, len(defs))
	for _, def := range defs {
		declared[def.Name] = true
	}
	for _, name := range slices.Sorted(maps.Keys(input)) {
		if !declared[name] {
			return nil, paramInvalid(name, "the template declares no such parameter")
		}
	}

	out := make(map[string]*Value, len(defs))
	for _, def := range defs {
		value, err := resolveParam(def, input[def.Name])
		if err != nil {
			return nil, err
		}
		out[def.Name] = value
	}
	return out, nil
}

// ValidateDefaults checks every declared default against its own parameter's
// constraints. The linter runs it; a default that fails its own pattern would
// otherwise surface as a form that cannot be submitted unchanged.
func ValidateDefaults(defs []*templatemodel.Parameter) error {
	for _, def := range defs {
		if isBlank(def.Default) {
			continue
		}
		if _, err := convertParam(def, def.Default); err != nil {
			return err
		}
	}
	return nil
}

func resolveParam(def *templatemodel.Parameter, raw any) (*Value, error) {
	if isBlank(raw) {
		raw = def.Default
	}
	if isBlank(raw) {
		switch {
		case def.Type == templatemodel.ParamTypeSecret && def.Generate != nil:
			secret, err := generateSecret(def.Generate)
			if err != nil {
				return nil, err
			}
			return &Value{Param: def, Value: secret, Generated: true}, nil
		case def.Optional:
			return &Value{Param: def}, nil
		default:
			return nil, paramInvalid(def.Name, "a value is required")
		}
	}

	value, err := convertParam(def, raw)
	if err != nil {
		return nil, err
	}
	return &Value{Param: def, Value: value}, nil
}

func isBlank(v any) bool {
	if v == nil {
		return true
	}
	text, isText := v.(string)
	return isText && text == ""
}

func convertParam(def *templatemodel.Parameter, raw any) (any, error) {
	switch def.Type {
	case templatemodel.ParamTypeString, templatemodel.ParamTypeSecret:
		text, ok := raw.(string)
		if !ok {
			return nil, paramInvalid(def.Name, "must be text")
		}
		return text, checkText(def, text)
	case templatemodel.ParamTypeInt:
		number, ok := toInt(raw)
		if !ok {
			return nil, paramInvalid(def.Name, "must be a whole number")
		}
		return number, checkInt(def, number)
	case templatemodel.ParamTypeSize:
		size, ok := toSize(raw)
		if !ok {
			return nil, paramInvalid(def.Name, "must be a size such as 512MB")
		}
		return size, checkSize(def, size)
	case templatemodel.ParamTypeBool:
		flag, ok := toBool(raw)
		if !ok {
			return nil, paramInvalid(def.Name, "must be true or false")
		}
		return flag, nil
	case templatemodel.ParamTypeSelect:
		text, ok := raw.(string)
		if !ok || !slices.ContainsFunc(def.Options, func(o *templatemodel.SelectOption) bool {
			return o.Value == text
		}) {
			return nil, paramInvalid(def.Name, "must be one of the template's options")
		}
		return text, nil
	case templatemodel.ParamTypeVolume:
		text, ok := raw.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil, paramInvalid(def.Name, "must name a volume")
		}
		return text, nil
	default:
		return nil, paramInvalid(def.Name, fmt.Sprintf("type %q is not supported", def.Type))
	}
}

func checkText(def *templatemodel.Parameter, text string) error {
	length := utf8.RuneCountInString(text)
	if def.MinLength != nil && length < *def.MinLength {
		return paramInvalid(def.Name, fmt.Sprintf("must be at least %d characters", *def.MinLength))
	}
	if def.MaxLength != nil && length > *def.MaxLength {
		return paramInvalid(def.Name, fmt.Sprintf("must be at most %d characters", *def.MaxLength))
	}
	if def.Pattern != "" {
		matched, err := regexp.MatchString(def.Pattern, text)
		if err != nil || !matched {
			return paramInvalid(def.Name, fmt.Sprintf("must match %s", def.Pattern))
		}
	}
	return nil
}

func checkInt(def *templatemodel.Parameter, number int64) error {
	minValue, maxValue := def.IntBounds()
	if minValue != nil && number < *minValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at least %d", *minValue))
	}
	if maxValue != nil && number > *maxValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at most %d", *maxValue))
	}
	return nil
}

func checkSize(def *templatemodel.Parameter, size unit.DataSize) error {
	minValue, maxValue := def.SizeBounds()
	if minValue != nil && size < *minValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at least %s", minValue.HR()))
	}
	if maxValue != nil && size > *maxValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at most %s", maxValue.HR()))
	}
	return nil
}

func toInt(raw any) (int64, bool) {
	if text, isText := raw.(string); isText {
		number, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		return number, err == nil
	}
	return templatemodel.ToInt64(raw)
}

func toSize(raw any) (unit.DataSize, bool) {
	text, isText := raw.(string)
	if !isText {
		return 0, false
	}
	size, err := unit.ParseDataSizeString(strings.TrimSpace(text))
	return size, err == nil
}

func toBool(raw any) (bool, bool) {
	switch v := raw.(type) {
	case bool:
		return v, true
	case string:
		flag, err := strconv.ParseBool(strings.TrimSpace(v))
		return flag, err == nil
	default:
		return false, false
	}
}

// generateSecret draws from crypto/rand. rand.Int rejects rather than takes a
// modulus, so every character of the alphabet is equally likely.
func generateSecret(gen *templatemodel.Generate) (string, error) {
	alphabet := alnumAlphabet
	if gen.Charset == templatemodel.CharsetHex {
		alphabet = hexAlphabet
	}
	limit := big.NewInt(int64(len(alphabet)))
	out := make([]byte, gen.Length)
	for i := range out {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}

func paramInvalid(name, reason string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).WithParam("Name", name).
		WithExtraDetail("%s: %s", name, reason)
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/`
Expected: PASS.

- [ ] **Step 6: Lint and commit**

Run: `make fmt && go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templaterender/params.go \
  hivepaas_app/service/apptemplateservice/templaterender/params_test.go \
  hivepaas_app/hperrors/errors_app_template.go hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml
git commit -m "feat(apptemplate): resolve template parameters and generate secrets"
```

---

## Task 5: Rendering - `templaterender.Render`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templaterender/render.go`
- Modify: `hivepaas_app/hperrors/errors_app_template.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`
- Test: `hivepaas_app/service/apptemplateservice/templaterender/render_test.go`

**Interfaces:**
- Consumes: `templatemodel.Template` and its lookups (Task 1); `specmodel.AppDoc`, `specmodel.CheckBuildable` (Task 2); `MergePatch`, `deepCopy`, `Substitute`, `Resolver` (Task 3); `ResolveParams`, `Value` (Task 4).
- Produces:
  - `type Request struct { Template *templatemodel.Template; Version, Variant string; Params map[string]any; AllowDeprecated bool }`
  - `type Result struct { Doc *specmodel.AppDoc; Params map[string]*Value; Version *templatemodel.Version; Variant *templatemodel.Variant; Image string; Base []byte; BaseSHA256 string }`
  - `func Render(req *Request) (*Result, error)`
  - `hperrors.ErrAppTemplateVersionNotFound`, `ErrAppTemplateVersionDeprecated`, `ErrAppTemplateVariantUnavailable`

`Version` and `Variant` empty mean the template's defaults. `Variant` is nil in the result for a template without variants. `Base` is canonical YAML of the render with secret placeholders left in; it never contains a secret.

- [ ] **Step 1: Declare the errors**

Add to `hivepaas_app/hperrors/errors_app_template.go`:

```go
	ErrAppTemplateVersionNotFound    = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_VERSION_NOT_FOUND")
	ErrAppTemplateVersionDeprecated  = NewErr(ErrNotAllowed, "ERR_APP_TEMPLATE_VERSION_DEPRECATED")
	ErrAppTemplateVariantUnavailable = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_VARIANT_UNAVAILABLE")
```

Add to `errors.app_template.en.toml`:

```toml
ERR_APP_TEMPLATE_VERSION_NOT_FOUND = "Version '{{.Version}}' of template '{{.Template}}' does not exist"
ERR_APP_TEMPLATE_VERSION_DEPRECATED = "Version '{{.Version}}' of template '{{.Template}}' is deprecated and cannot be used for a new app"
ERR_APP_TEMPLATE_VARIANT_UNAVAILABLE = "Variant '{{.Variant}}' is not available for version '{{.Version}}' of template '{{.Template}}'"
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/apptemplateservice/templaterender/render_test.go`:

```go
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
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/ -run Render`
Expected: FAIL - `undefined: Render`.

- [ ] **Step 4: Write the renderer**

`hivepaas_app/service/apptemplateservice/templaterender/render.go`:

```go
package templaterender

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const yamlIndent = 2

type Request struct {
	Template *templatemodel.Template
	// Version and Variant are the user's choice; empty means the template's default.
	Version string
	Variant string
	Params  map[string]any
	// AllowDeprecated renders a deprecated version, which creating an app never
	// does. The linter and phase 2's updates of existing apps do.
	AllowDeprecated bool
}

type Result struct {
	// Doc is what the app is built from, secrets included.
	Doc     *specmodel.AppDoc
	Params  map[string]*Value
	Version *templatemodel.Version
	// Variant is nil for a template without variants.
	Variant *templatemodel.Variant
	Image   string

	// Base is the same render with every secret placeholder left in, as canonical
	// YAML: the base phase 2's three-way merge compares against. It is stored, so
	// it must never hold a secret, and it is hashed, so it must be byte-stable.
	Base       []byte
	BaseSHA256 string
}

// Render turns a template and a user's choices into an AppDoc.
//
// Every step is pure, in this order: validate, choose version and variant,
// resolve parameters, apply the version override, substitute placeholders,
// decode strictly as an AppDoc, refuse what cannot be built. The base render
// repeats the substitution with secrets kept as placeholders.
func Render(req *Request) (*Result, error) {
	tmpl := req.Template
	if err := tmpl.Validate(""); err != nil {
		return nil, hperrors.Wrap(err)
	}
	version, variant, image, err := selectVersion(tmpl, req)
	if err != nil {
		return nil, err
	}
	params, err := ResolveParams(tmpl.Parameters, req.Params)
	if err != nil {
		return nil, err
	}

	tree := deepCopy(tmpl.App)
	if version.Override != nil {
		// TODO: app templates phase 3 - apply the variant's override after the
		// version's. See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
		tree = MergePatch(tree, version.Override.App)
	}
	vars := templateVars(version, variant, image)
	name := tmpl.Metadata.Name

	applied, err := Substitute(tree, resolver(name, vars, params, false))
	if err != nil {
		return nil, err
	}
	doc, err := decodeAppDoc(name, applied)
	if err != nil {
		return nil, err
	}
	if err = specmodel.CheckBuildable(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}

	baseTree, err := Substitute(tree, resolver(name, vars, params, true))
	if err != nil {
		return nil, err
	}
	base, err := marshalCanonical(baseTree)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(base)

	return &Result{
		Doc:        doc,
		Params:     params,
		Version:    version,
		Variant:    variant,
		Image:      image,
		Base:       base,
		BaseSHA256: hex.EncodeToString(sum[:]),
	}, nil
}

func selectVersion(tmpl *templatemodel.Template, req *Request) (
	*templatemodel.Version, *templatemodel.Variant, string, error) {
	name := tmpl.Metadata.Name
	version := tmpl.DefaultVersion()
	if req.Version != "" {
		version = tmpl.FindVersion(req.Version)
	}
	if version == nil {
		return nil, nil, "", hperrors.Wrap(hperrors.ErrAppTemplateVersionNotFound).
			WithParam("Template", name).WithParam("Version", req.Version)
	}
	if version.Deprecated && !req.AllowDeprecated {
		return nil, nil, "", hperrors.Wrap(hperrors.ErrAppTemplateVersionDeprecated).
			WithParam("Template", name).WithParam("Version", version.Name)
	}

	if len(tmpl.Variants) == 0 {
		if req.Variant != "" {
			return nil, nil, "", variantUnavailable(name, version.Name, req.Variant)
		}
		return version, nil, version.Image, nil
	}

	variant := tmpl.DefaultVariant()
	if req.Variant != "" {
		variant = tmpl.FindVariant(req.Variant)
	}
	if variant == nil || version.ImageFor(variant.Name) == "" {
		wanted := req.Variant
		if wanted == "" && variant != nil {
			wanted = variant.Name
		}
		return nil, nil, "", variantUnavailable(name, version.Name, wanted)
	}
	return version, variant, version.ImageFor(variant.Name), nil
}

func variantUnavailable(template, version, variant string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateVariantUnavailable).
		WithParam("Template", template).WithParam("Version", version).WithParam("Variant", variant)
}

// templateVars are the placeholders a template can use besides its parameters.
//
// TODO: app templates phase 3 - an app namespace (app.name, app.key). See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func templateVars(version *templatemodel.Version, variant *templatemodel.Variant, image string) map[string]any {
	vars := map[string]any{
		"version.name":    version.Name,
		"version.release": version.Release,
		"variant.name":    "",
		"image":           image,
	}
	if variant != nil {
		vars["variant.name"] = variant.Name
	}
	for key, value := range version.Vars {
		vars["version.vars."+key] = value
	}
	return vars
}

// resolver answers placeholders from the version, the variant and the resolved
// parameters. keepSecrets leaves secret placeholders as they are - the base render.
func resolver(template string, vars map[string]any, params map[string]*Value, keepSecrets bool) Resolver {
	return func(ref string) (any, bool, error) {
		if paramName, isParam := strings.CutPrefix(ref, "params."); isParam {
			value, found := params[paramName]
			if !found {
				return nil, false, undefinedPlaceholder(template, ref)
			}
			if keepSecrets && value.Param.Type == templatemodel.ParamTypeSecret {
				return nil, true, nil
			}
			if value.Value == nil {
				return "", false, nil
			}
			return value.Value, false, nil
		}
		value, found := vars[ref]
		if !found {
			return nil, false, undefinedPlaceholder(template, ref)
		}
		return value, false, nil
	}
}

func undefinedPlaceholder(template, ref string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
		WithExtraDetail("%s: placeholder %q refers to nothing", template, ref)
}

func decodeAppDoc(template string, tree any) (*specmodel.AppDoc, error) {
	encoded, err := yaml.Marshal(tree)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	doc := &specmodel.AppDoc{}
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	decoder.KnownFields(true)
	if err = decoder.Decode(doc); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s: app: %s", template, err.Error())
	}
	return doc, nil
}

// marshalCanonical writes a tree as YAML. yaml.v3 sorts map keys, which is what
// makes the same render hash the same every time.
func marshalCanonical(tree any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(yamlIndent)
	if err := encoder.Encode(tree); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := encoder.Close(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/`
Expected: PASS.

- [ ] **Step 6: Lint and commit**

Run: `make fmt && go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templaterender/render.go \
  hivepaas_app/service/apptemplateservice/templaterender/render_test.go \
  hivepaas_app/hperrors/errors_app_template.go hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml
git commit -m "feat(apptemplate): render a template into an AppDoc with a secret-free base"
```

---
## Task 6: Repository tooling - `templaterepo`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templaterepo/repo.go`
- Create: `hivepaas_app/service/apptemplateservice/templaterepo/lint.go`
- Create: `hivepaas_app/service/apptemplateservice/templaterepo/index.go`
- Modify: `hivepaas_app/hperrors/errors_app_template.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`
- Test: `hivepaas_app/service/apptemplateservice/templaterepo/repo_test.go`

**Interfaces:**
- Consumes: `templatemodel` decoders, `Validate`, `Index` types (Task 1); `templaterender.Render`, `Request`, `ValidateDefaults` (Tasks 4-5).
- Produces:
  - constants `CategoriesFile = "categories.yaml"`, `TagsFile = "tags.yaml"`, `IndexFile = "index.json"`, `TemplatesDir = "templates"`, `MaxIndexSize = 1 << 20`, `MaxTemplateSize = 256 << 10`, `MaxIconSize = 256 << 10`
  - `type File struct { Path string; Content []byte; SHA256 string }`
  - `type TemplateFile struct { File; Template *templatemodel.Template }`
  - `type Repo struct { Categories *templatemodel.Categories; Tags *templatemodel.Tags; Templates []*TemplateFile; Icons map[string]*File }`
  - `type Problem struct { Path, Message string }` with `String()`
  - `func Load(fsys fs.FS) (*Repo, []Problem, error)` - an error only when the vocabularies cannot be read; a broken template is a problem, not an error
  - `func Lint(repo *Repo) []Problem`
  - `func BuildIndex(repo *Repo) (*templatemodel.Index, error)`, `func MarshalIndex(index *templatemodel.Index) ([]byte, error)`
  - `func (r *Repo) FindTemplate(name string) *TemplateFile`
  - `hperrors.ErrAppTemplateFileTooLarge`

- [ ] **Step 1: Declare the error**

Add to `hivepaas_app/hperrors/errors_app_template.go`:

```go
	ErrAppTemplateFileTooLarge = NewErr(ErrTooBig, "ERR_APP_TEMPLATE_FILE_TOO_LARGE")
```

Add to `errors.app_template.en.toml`:

```toml
ERR_APP_TEMPLATE_FILE_TOO_LARGE = "An app template file is larger than allowed"
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/apptemplateservice/templaterepo/repo_test.go`:

```go
package templaterepo

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const categoriesYAML = `
apiVersion: hivepaas.com/v1
kind: TemplateCategories
categories:
  - id: databases
    title: Databases
    children:
      - {id: sql, title: SQL}
      - {id: cache, title: Cache & Queues}
`

const tagsYAML = `
apiVersion: hivepaas.com/v1
kind: TemplateTags
tags:
  - {id: sql, title: SQL}
`

const demoTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: demo
  title: Demo
  tagline: A demo database
  description: Demo.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
parameters:
  - {name: password, title: Password, type: secret, generate: {length: 16}}
  - {name: dataVolume, title: Data volume, type: volume}
versions:
  - {name: "2", release: "2.1", default: true, image: "demo:2.1.0"}
  - {name: "1", release: "1.9", deprecated: true, image: "demo:1.9.3"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /data: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    kind: {category: database, engine: demo, database: {password: "${{ params.password }}"}}
`

const demoIcon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"/>`

func validRepoFS() fstest.MapFS {
	return fstest.MapFS{
		CategoriesFile:        {Data: []byte(categoriesYAML)},
		TagsFile:              {Data: []byte(tagsYAML)},
		"templates/demo.yaml": {Data: []byte(demoTemplateYAML)},
		"icons/demo.svg":      {Data: []byte(demoIcon)},
	}
}

func withFile(fsys fstest.MapFS, path, content string) fstest.MapFS {
	out := maps.Clone(fsys)
	out[path] = &fstest.MapFile{Data: []byte(content)}
	return out
}

func loadAndLint(t *testing.T, fsys fstest.MapFS) (*Repo, []Problem) {
	t.Helper()
	repo, problems, err := Load(fsys)
	assert.NoError(t, err)
	return repo, append(problems, Lint(repo)...)
}

func TestLoadAndLintAValidRepository(t *testing.T) {
	repo, problems := loadAndLint(t, validRepoFS())

	assert.Empty(t, problems)
	assert.Len(t, repo.Templates, 1)
	assert.Equal(t, "demo", repo.FindTemplate("demo").Template.Metadata.Name)
	assert.Equal(t, demoIcon, string(repo.Icons["icons/demo.svg"].Content))
}

func TestLintFindsProblems(t *testing.T) {
	cases := map[string]struct {
		fsys fstest.MapFS
		want string
	}{
		"unknown category": {
			withFile(validRepoFS(), "templates/demo.yaml",
				strings.Replace(demoTemplateYAML, "databases/sql", "databases/graph", 1)),
			`category "databases/graph" is not in categories.yaml`,
		},
		"unknown tag": {
			withFile(validRepoFS(), "templates/demo.yaml", strings.Replace(demoTemplateYAML, "[sql]", "[nosql]", 1)),
			`tag "nosql" is not in tags.yaml`,
		},
		"missing icon": {
			func() fstest.MapFS { fsys := validRepoFS(); delete(fsys, "icons/demo.svg"); return fsys }(),
			"icons/demo.svg",
		},
		"floating image": {
			withFile(validRepoFS(), "templates/demo.yaml", strings.Replace(demoTemplateYAML, "demo:2.1.0", "demo:2", 1)),
			"exact release",
		},
		"undefined placeholder": {
			withFile(validRepoFS(), "templates/demo.yaml",
				strings.Replace(demoTemplateYAML, "params.dataVolume", "params.volume", 1)),
			`placeholder "params.volume" refers to nothing`,
		},
		"unknown field": {
			withFile(validRepoFS(), "templates/demo.yaml", demoTemplateYAML+"\nextra: 1\n"),
			"extra",
		},
		"file name mismatch": {
			withFile(validRepoFS(), "templates/other.yaml", demoTemplateYAML),
			`must equal the file name "other"`,
		},
		"oversized template": {
			withFile(validRepoFS(), "templates/huge.yaml", strings.Repeat("#", MaxTemplateSize+1)),
			"templates/huge.yaml",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, problems := loadAndLint(t, tc.fsys)
			var all []string
			for _, problem := range problems {
				all = append(all, problem.String())
			}
			assert.Contains(t, strings.Join(all, "\n"), tc.want)
		})
	}
}

func TestLoadRequiresTheVocabularies(t *testing.T) {
	fsys := validRepoFS()
	delete(fsys, CategoriesFile)
	_, _, err := Load(fsys)
	assert.Error(t, err)
}

func TestBuildIndex(t *testing.T) {
	repo, problems := loadAndLint(t, validRepoFS())
	assert.Empty(t, problems)

	index, err := BuildIndex(repo)
	assert.NoError(t, err)

	entry := index.FindTemplate("demo")
	templateSum := sha256.Sum256([]byte(demoTemplateYAML))
	iconSum := sha256.Sum256([]byte(demoIcon))
	assert.Equal(t, templatemodel.FileRef{Path: "templates/demo.yaml", SHA256: hex.EncodeToString(templateSum[:])},
		entry.File)
	assert.Equal(t, templatemodel.FileRef{Path: "icons/demo.svg", SHA256: hex.EncodeToString(iconSum[:])}, entry.Icon)
	assert.Equal(t, []*templatemodel.IndexVersion{
		{Name: "2", Release: "2.1", Default: true},
		{Name: "1", Release: "1.9", Deprecated: true},
	}, entry.Versions)
	assert.Equal(t, "Cache & Queues", index.Categories[0].Children[1].Title)

	data, err := MarshalIndex(index)
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(string(data), "}\n"))
	assert.Contains(t, string(data), "Cache & Queues", "no HTML escaping in a file people read")

	again, err := MarshalIndex(index)
	assert.NoError(t, err)
	assert.Equal(t, data, again)

	decoded, err := templatemodel.DecodeIndex(data)
	assert.NoError(t, err)
	assert.Equal(t, index, decoded)
}
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterepo/`
Expected: FAIL - `undefined: Load`.

- [ ] **Step 4: Write the loader**

`hivepaas_app/service/apptemplateservice/templaterepo/repo.go`:

```go
// Package templaterepo reads a checkout of an app templates repository: it
// loads the files, lints them and builds index.json. It works over an fs.FS and
// nothing else, so tools/apptemplate and the development source read a checkout
// exactly the same way.
package templaterepo

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	CategoriesFile = "categories.yaml"
	TagsFile       = "tags.yaml"
	IndexFile      = "index.json"
	TemplatesDir   = "templates"

	MaxIndexSize    = 1 << 20
	MaxTemplateSize = 256 << 10
	MaxIconSize     = 256 << 10
)

type File struct {
	Path    string
	Content []byte
	SHA256  string
}

type TemplateFile struct {
	File
	Template *templatemodel.Template
}

type Repo struct {
	Categories *templatemodel.Categories
	Tags       *templatemodel.Tags
	// Templates are sorted by name.
	Templates []*TemplateFile
	// Icons holds the icons templates name, by path.
	Icons map[string]*File
}

// Problem is one thing wrong with a repository, for a person to fix.
type Problem struct {
	Path    string
	Message string
}

func (p Problem) String() string {
	return p.Path + ": " + p.Message
}

func (r *Repo) FindTemplate(name string) *TemplateFile {
	for _, file := range r.Templates {
		if file.Template.Metadata.Name == name {
			return file
		}
	}
	return nil
}

// Load reads a checkout. The two vocabularies are required, so failing to read
// either is an error. A template that cannot be read or decoded is a problem
// rather than an error: one broken file must not hide what is wrong with the rest.
func Load(fsys fs.FS) (*Repo, []Problem, error) {
	categoriesData, err := readLimited(fsys, CategoriesFile, MaxTemplateSize)
	if err != nil {
		return nil, nil, err
	}
	categories, err := templatemodel.DecodeCategories(categoriesData)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	tagsData, err := readLimited(fsys, TagsFile, MaxTemplateSize)
	if err != nil {
		return nil, nil, err
	}
	tags, err := templatemodel.DecodeTags(tagsData)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	repo := &Repo{Categories: categories, Tags: tags, Icons: map[string]*File{}}
	var problems []Problem

	paths, err := fs.Glob(fsys, TemplatesDir+"/*.yaml")
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	for _, path := range paths {
		data, err := readLimited(fsys, path, MaxTemplateSize)
		if err != nil {
			problems = append(problems, Problem{Path: path, Message: ErrorText(err)})
			continue
		}
		tmpl, err := templatemodel.DecodeTemplate(data)
		if err != nil {
			problems = append(problems, Problem{Path: path, Message: ErrorText(err)})
			continue
		}
		repo.Templates = append(repo.Templates, &TemplateFile{File: newFile(path, data), Template: tmpl})
		problems = append(problems, repo.loadIcon(fsys, tmpl.Metadata.Icon)...)
	}

	sort.Slice(repo.Templates, func(i, j int) bool {
		return repo.Templates[i].Template.Metadata.Name < repo.Templates[j].Template.Metadata.Name
	})
	return repo, problems, nil
}

func (r *Repo) loadIcon(fsys fs.FS, path string) []Problem {
	if path == "" || r.Icons[path] != nil {
		return nil
	}
	if !fs.ValidPath(path) {
		return []Problem{{Path: path, Message: "an icon path must stay inside the repository"}}
	}
	data, err := readLimited(fsys, path, MaxIconSize)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // Lint reports a missing icon against the template that names it
		}
		return []Problem{{Path: path, Message: ErrorText(err)}}
	}
	r.Icons[path] = newFile(path, data)
	return nil
}

func newFile(path string, data []byte) *File {
	sum := sha256.Sum256(data)
	return &File{Path: path, Content: data, SHA256: hex.EncodeToString(sum[:])}
}

func readLimited(fsys fs.FS, name string, limit int64) ([]byte, error) {
	file, err := fsys.Open(name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if int64(len(data)) > limit {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateFileTooLarge).
			WithExtraDetail("%s is larger than %d bytes", name, limit)
	}
	return data, nil
}

// ErrorText renders an error for a person reading lint output: the translated
// message and its detail, rather than an error code.
func ErrorText(err error) string {
	var hpErr hperrors.HPError
	if errors.As(err, &hpErr) {
		return strings.Join(strings.Fields(hpErr.Build("en").Detail), " ")
	}
	return err.Error()
}
```

- [ ] **Step 5: Write the linter**

`hivepaas_app/service/apptemplateservice/templaterepo/lint.go`:

```go
package templaterepo

import (
	"fmt"
	"path"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// lintVolumeID and lintSecret stand in for what a user supplies, so a template
// can be rendered without one.
const (
	lintVolumeID = "lint-volume"
	lintSecret   = "lint-secret-value"
)

// Lint checks what Load could read: each template on its own, against the
// vocabularies and its icon, and by rendering every version and variant with
// default parameters - the same render HivePaaS runs.
func Lint(repo *Repo) []Problem {
	problems := lintVocabularies(repo)
	categories := categoryRefs(repo.Categories.Categories)
	tags := map[string]bool{}
	for _, tag := range repo.Tags.Tags {
		tags[tag.ID] = true
	}

	for _, file := range repo.Templates {
		tmpl := file.Template
		report := func(format string, args ...any) {
			problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(format, args...)})
		}

		name := strings.TrimSuffix(path.Base(file.Path), ".yaml")
		if err := tmpl.Validate(name); err != nil {
			report("%s", ErrorText(err))
			continue
		}
		for _, category := range tmpl.Metadata.Categories {
			if !categories[category] {
				report("category %q is not in %s", category, CategoriesFile)
			}
		}
		for _, tag := range tmpl.Metadata.Tags {
			if !tags[tag] {
				report("tag %q is not in %s", tag, TagsFile)
			}
		}
		if repo.Icons[tmpl.Metadata.Icon] == nil {
			report("icon %q is missing", tmpl.Metadata.Icon)
		}
		if err := templaterender.ValidateDefaults(tmpl.Parameters); err != nil {
			report("%s", ErrorText(err))
			continue
		}
		problems = append(problems, lintRenders(file)...)
	}
	return problems
}

func lintVocabularies(repo *Repo) []Problem {
	var problems []Problem
	parents := map[string]bool{}
	for _, category := range repo.Categories.Categories {
		if parents[category.ID] {
			problems = append(problems, Problem{Path: CategoriesFile,
				Message: fmt.Sprintf("category %q is declared twice", category.ID)})
		}
		parents[category.ID] = true
		if len(category.Children) == 0 {
			problems = append(problems, Problem{Path: CategoriesFile,
				Message: fmt.Sprintf("category %q has no children to put templates in", category.ID)})
		}
	}
	tags := map[string]bool{}
	for _, tag := range repo.Tags.Tags {
		if tags[tag.ID] {
			problems = append(problems, Problem{Path: TagsFile, Message: fmt.Sprintf("tag %q is declared twice", tag.ID)})
		}
		tags[tag.ID] = true
	}
	return problems
}

func categoryRefs(categories []*templatemodel.Category) map[string]bool {
	refs := map[string]bool{}
	for _, parent := range categories {
		for _, child := range parent.Children {
			refs[parent.ID+"/"+child.ID] = true
		}
	}
	return refs
}

func lintRenders(file *TemplateFile) []Problem {
	tmpl := file.Template
	variants := []string{""}
	if len(tmpl.Variants) > 0 {
		variants = variants[:0]
		for _, variant := range tmpl.Variants {
			variants = append(variants, variant.Name)
		}
	}

	var problems []Problem
	for _, version := range tmpl.Versions {
		for _, variant := range variants {
			if variant != "" && version.ImageFor(variant) == "" {
				continue
			}
			_, err := templaterender.Render(&templaterender.Request{
				Template:        tmpl,
				Version:         version.Name,
				Variant:         variant,
				Params:          lintParams(tmpl),
				AllowDeprecated: true,
			})
			if err != nil {
				label := "version " + version.Name
				if variant != "" {
					label += ", variant " + variant
				}
				problems = append(problems, Problem{Path: file.Path, Message: label + ": " + ErrorText(err)})
			}
		}
	}
	return problems
}

// lintParams fills in what defaults cannot: a volume, and a secret that has no
// generate. Everything else renders from its default, and a required parameter
// without one fails the render - which is the point.
func lintParams(tmpl *templatemodel.Template) map[string]any {
	params := map[string]any{}
	for _, param := range tmpl.Parameters {
		switch param.Type {
		case templatemodel.ParamTypeVolume:
			params[param.Name] = lintVolumeID
		case templatemodel.ParamTypeSecret:
			if param.Generate == nil {
				params[param.Name] = lintSecret
			}
		case templatemodel.ParamTypeString, templatemodel.ParamTypeInt, templatemodel.ParamTypeSize,
			templatemodel.ParamTypeBool, templatemodel.ParamTypeSelect:
		}
	}
	return params
}
```

- [ ] **Step 6: Write the index builder**

`hivepaas_app/service/apptemplateservice/templaterepo/index.go`:

```go
package templaterepo

import (
	"bytes"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const indexIndent = "  "

// BuildIndex builds index.json from a loaded repository. It assumes Lint found
// nothing; the one thing it cannot build around is a missing icon, since the
// icon's hash is part of every entry.
func BuildIndex(repo *Repo) (*templatemodel.Index, error) {
	index := &templatemodel.Index{
		APIVersion: templatemodel.APIVersion,
		Kind:       templatemodel.KindTemplateIndex,
		Categories: repo.Categories.Categories,
		Tags:       repo.Tags.Tags,
		Templates:  []*templatemodel.IndexEntry{},
	}
	for _, file := range repo.Templates {
		tmpl := file.Template
		icon := repo.Icons[tmpl.Metadata.Icon]
		if icon == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: icon %q is missing", file.Path, tmpl.Metadata.Icon)
		}

		entry := &templatemodel.IndexEntry{
			Name:       tmpl.Metadata.Name,
			File:       templatemodel.FileRef{Path: file.Path, SHA256: file.SHA256},
			Icon:       templatemodel.FileRef{Path: icon.Path, SHA256: icon.SHA256},
			Title:      tmpl.Metadata.Title,
			Tagline:    tmpl.Metadata.Tagline,
			Categories: tmpl.Metadata.Categories,
			Tags:       tmpl.Metadata.Tags,
			Aliases:    tmpl.Metadata.Aliases,
			Requires:   tmpl.Metadata.Requires,
		}
		for _, variant := range tmpl.Variants {
			entry.Variants = append(entry.Variants,
				&templatemodel.IndexVariant{Name: variant.Name, Default: variant.Default})
		}
		for _, version := range tmpl.Versions {
			indexVersion := &templatemodel.IndexVersion{
				Name:       version.Name,
				Release:    version.Release,
				Default:    version.Default,
				Deprecated: version.Deprecated,
			}
			for _, variant := range tmpl.Variants {
				if version.ImageFor(variant.Name) != "" {
					indexVersion.Variants = append(indexVersion.Variants, variant.Name)
				}
			}
			entry.Versions = append(entry.Versions, indexVersion)
		}
		index.Templates = append(index.Templates, entry)
	}
	return index, nil
}

// MarshalIndex writes index.json: indented, without HTML escaping, ending in a
// newline, and byte-identical for the same repository.
func MarshalIndex(index *templatemodel.Index) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indexIndent)
	if err := encoder.Encode(index); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 7: Run the tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterepo/`
Expected: PASS.

- [ ] **Step 8: Lint and commit**

Run: `make fmt && go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templaterepo \
  hivepaas_app/hperrors/errors_app_template.go hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml
git commit -m "feat(apptemplate): load, lint and index a templates repository"
```

---

## Task 7: The `app-template` setting

**Files:**
- Modify: `hivepaas_app/base/setting.go`
- Create: `hivepaas_app/entity/setting_app_template.go`
- Create: `hivepaas_app/entity/setting_app_template_migration.go`
- Modify: `hivepaas_app/entity/setting_spec.go`
- Modify: `hivepaas_app/service/specservice/specmodel/singleton.go`
- Test: `hivepaas_app/entity/setting_app_template_test.go`

**Interfaces:**
- Consumes: `entity.EncryptedField`, `registerSettingParser`, `registerDefaultSpecPolicies`.
- Produces:
  - `base.SettingTypeAppTemplate SettingType = "app-template"`
  - `entity.CurrentAppTemplateSettingsVersion = 1`
  - `type AppTemplateSettings struct { Source, Template, Title, Version, Variant string; Params map[string]*AppTemplateParam; Base AppTemplateBase }`
  - `type AppTemplateParam struct { Value string; Secret EncryptedField }`
  - `type AppTemplateBase struct { Revision, Release, TemplateSHA256, Rendered, RenderedSHA256 string; AppliedAt time.Time }`
  - `func (s *Setting) AsAppTemplateSettings() (*AppTemplateSettings, error)`, `MustAsAppTemplateSettings()`
  - block name `template` in spec documents

`Title` is stored so reading an app's binding needs neither the network nor the index.

- [ ] **Step 1: Write the failing test**

`hivepaas_app/entity/setting_app_template_test.go`:

```go
package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func newTestAppTemplateSettings() *AppTemplateSettings {
	return &AppTemplateSettings{
		Source:   "official",
		Template: "postgres",
		Title:    "PostgreSQL",
		Version:  "17",
		Variant:  "alpine",
		Params: map[string]*AppTemplateParam{
			"username": {Value: "app"},
			"password": {Secret: NewEncryptedField("generated-password")},
		},
		Base: AppTemplateBase{
			Revision:       "0123456789abcdef0123456789abcdef01234567",
			Release:        "17.6",
			TemplateSHA256: "aa",
			Rendered:       "deployment: {}\n",
			RenderedSHA256: "bb",
			AppliedAt:      time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC),
		},
	}
}

func TestAppTemplateSettingsRoundTrip(t *testing.T) {
	useDataKey(t)
	setting := &Setting{Type: base.SettingTypeAppTemplate}

	assert.NoError(t, setting.SetData(newTestAppTemplateSettings()))
	assert.NotContains(t, setting.Data, "generated-password", "a secret parameter is stored encrypted")

	restored := &Setting{Type: base.SettingTypeAppTemplate, Data: setting.Data}
	parsed, err := restored.AsAppTemplateSettings()
	assert.NoError(t, err)
	assert.Equal(t, "app", parsed.Params["username"].Value)
	assert.NoError(t, parsed.Decrypt())
	password, err := parsed.Params["password"].Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "generated-password", password)
	assert.Equal(t, "17.6", parsed.Base.Release)
}

func TestAppTemplateSettingsOmitsSecrets(t *testing.T) {
	useDataKey(t)
	settings := newTestAppTemplateSettings()

	count, err := OmitSecrets(settings)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.True(t, settings.Params["password"].Secret.IsEmpty())
	encoded, err := json.Marshal(settings)
	assert.NoError(t, err)
	assert.NotContains(t, string(encoded), `"secret"`, "an empty secret is left out")
}
```

- [ ] **Step 2: Run it to see it fail**

Run: `go test ./hivepaas_app/entity/ -run AppTemplate`
Expected: FAIL - `undefined: AppTemplateSettings`.

- [ ] **Step 3: Register the type**

In `hivepaas_app/base/setting.go`, add after `SettingTypeAppRouting`:

```go
	SettingTypeAppTemplate       SettingType = "app-template"
```

`hivepaas_app/entity/setting_app_template.go`:

```go
package entity

import (
	"maps"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	CurrentAppTemplateSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppTemplate, &appTemplateSettingsParser{})

type appTemplateSettingsParser struct {
}

func (s *appTemplateSettingsParser) New() SettingData {
	return &AppTemplateSettings{}
}

// AppTemplateSettings records the template an app was provisioned from.
//
// Phase 1 writes it and reads it only for display. It is complete now so that
// phase 2 can update the apps phase 1 created without a migration that would
// have to reconstruct a base nobody recorded.
//
// TODO: app templates phase 2 - UpdatePolicy ("follow" or "pinned"). See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type AppTemplateSettings struct {
	Source   string `json:"source"`
	Template string `json:"template"`
	Title    string `json:"title"`
	// Version is the major line, such as "17".
	Version string `json:"version"`
	Variant string `json:"variant,omitempty"`

	// Params are the values given at creation, secrets encrypted. Parameters are
	// fixed after creation: a change is made on the app, and phase 2's merge sees
	// it as a user change.
	Params map[string]*AppTemplateParam `json:"params,omitempty"`

	Base AppTemplateBase `json:"base"`
}

type AppTemplateParam struct {
	Value  string         `json:"value,omitempty"`
	Secret EncryptedField `json:"secret,omitzero"`
}

// AppTemplateBase is what the template rendered to when it was last applied:
// the base of phase 2's three-way merge.
//
// It is the render, not the template file. A later HivePaaS that fixes a bug in
// rendering would render the old template differently from what was applied,
// and every app would show changes no template made. Rendered holds secret
// placeholders, never secrets.
type AppTemplateBase struct {
	Revision       string    `json:"revision"`
	Release        string    `json:"release"`
	TemplateSHA256 string    `json:"templateSha256"`
	Rendered       string    `json:"rendered"`
	RenderedSHA256 string    `json:"renderedSha256"`
	AppliedAt      time.Time `json:"appliedAt"`
}

func (s *AppTemplateSettings) GetType() base.SettingType {
	return base.SettingTypeAppTemplate
}

// GetRefObjectIDs is empty on purpose. A volume parameter records the id the
// user chose; the reference itself lives in the app's mount, where the storage
// settings already track it.
func (s *AppTemplateSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppTemplateSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// Decrypt implements the secret decrypter the settings reveal path looks for.
func (s *AppTemplateSettings) Decrypt() error {
	for _, name := range slices.Sorted(maps.Keys(s.Params)) {
		param := s.Params[name]
		if param == nil || param.Secret.IsEmpty() {
			continue
		}
		if _, err := param.Secret.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Setting) AsAppTemplateSettings() (*AppTemplateSettings, error) {
	return parseSettingAs[*AppTemplateSettings](s)
}

func (s *Setting) MustAsAppTemplateSettings() *AppTemplateSettings {
	return gofn.Must(s.AsAppTemplateSettings())
}
```

`hivepaas_app/entity/setting_app_template_migration.go`:

```go
package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"

func (s *AppTemplateSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppTemplateSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppTemplateSettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	setting.Version = CurrentAppTemplateSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
```

In `hivepaas_app/entity/setting_spec.go`, add to the `registerDefaultSpecPolicies(...)` list after `base.SettingTypeAppPlacement,`:

```go
		base.SettingTypeAppTemplate,
```

In `hivepaas_app/service/specservice/specmodel/singleton.go`, add to `singletonBlockNames` after the `base.SettingTypeAppRouting` line:

```go
	base.SettingTypeAppTemplate:   "template",
```

- [ ] **Step 4: Run the new test and the registries' completeness tests**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/service/specservice/...`
Expected: PASS - including `TestEverySettingTypeHasASpecPolicy` and `TestEverySettingTypeIsClassifiedExactlyOnce`, which fail if either registration above is missing.

- [ ] **Step 5: Lint and commit**

Run: `make fmt && go build ./... && make lint-local`
Expected: `0 issues.` If `exhaustive` reports a switch over `base.SettingType` missing the new value, add `base.SettingTypeAppTemplate` to the case that treats app-scope settings with no special behavior.

```bash
git add hivepaas_app/base/setting.go hivepaas_app/entity/setting_app_template.go \
  hivepaas_app/entity/setting_app_template_migration.go hivepaas_app/entity/setting_app_template_test.go \
  hivepaas_app/entity/setting_spec.go hivepaas_app/service/specservice/specmodel/singleton.go
git commit -m "feat(entity): app-template setting recording the template an app came from"
```

---
## Task 8: Move mount building into `volumeservice`

A usecase cannot call another usecase, and provisioning from a template (Task 9) has to build mounts exactly the way the storage settings screen does - subpath, bind rewrite, driver config, permissions, pin conflicts. That code moves from `appsettingsuc` into `volumeservice`, behind a request type of its own because a service cannot take `appsettingsdto.Mount`. Behavior does not change; the existing tests move with the code.

**Files:**
- Create: `hivepaas_app/service/volumeservice/types.go`
- Modify: `hivepaas_app/service/volumeservice/service.go`
- Move: `hivepaas_app/usecase/appsettingsuc/volume_mount.go` → `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_helpers.go`
- Move: `hivepaas_app/usecase/appsettingsuc/volume_mount_test.go` → `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_helpers_test.go`
- Delete: `hivepaas_app/usecase/appsettingsuc/volume_mount_wiring_test.go` (replaced by `app_mounts_test.go`)
- Create: `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mounts.go`
- Modify: `hivepaas_app/service/volumeservice/volumeserviceimpl/service.go`
- Modify: `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go`
- Test: `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mounts_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces:
  - `type volumeservice.AppMountReq struct { Type mount.Type; Source string; Target string; ReadOnly bool; Consistency mount.Consistency; VolumeOptions *AppMountVolumeOptions; ClusterOptions *AppMountVolumeOptions }` - `Source` is a cluster-volume setting id
  - `type volumeservice.AppMountVolumeOptions struct { Subpath string; NoCopy bool; Labels map[string]string; DriverConfig *mount.Driver }`
  - `type volumeservice.BuildAppMountsReq struct { App *entity.App; Kept []mount.Mount; New []*AppMountReq }`
  - `type volumeservice.BuildAppMountsResp struct { Mounts []mount.Mount }`
  - `volumeservice.Service.BuildAppMounts(ctx context.Context, db database.IDB, req *BuildAppMountsReq) (*BuildAppMountsResp, error)`

- [ ] **Step 1: Move the pure helpers and their tests**

```bash
git mv hivepaas_app/usecase/appsettingsuc/volume_mount.go \
  hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_helpers.go
git mv hivepaas_app/usecase/appsettingsuc/volume_mount_test.go \
  hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_helpers_test.go
git rm hivepaas_app/usecase/appsettingsuc/volume_mount_wiring_test.go
sed -i '' 's/^package appsettingsuc$/package volumeserviceimpl/' \
  hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_helpers.go \
  hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_helpers_test.go
```

Then move `getConfiguredPropagation` - the last function in `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go`, starting at `func getConfiguredPropagation(o string) mount.Propagation {` - to the end of `app_mount_helpers.go`, unchanged, and add `"strings"` to that file's imports.

- [ ] **Step 2: Declare the request types and the method**

`hivepaas_app/service/volumeservice/types.go`:

```go
package volumeservice

import (
	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// AppMountReq asks for one cluster-volume setting to be mounted into an app.
type AppMountReq struct {
	Type mount.Type
	// Source is the id of the cluster-volume setting, not a docker volume name.
	Source         string
	Target         string
	ReadOnly       bool
	Consistency    mount.Consistency
	VolumeOptions  *AppMountVolumeOptions
	ClusterOptions *AppMountVolumeOptions
}

type AppMountVolumeOptions struct {
	Subpath      string
	NoCopy       bool
	Labels       map[string]string
	DriverConfig *mount.Driver
}

type BuildAppMountsReq struct {
	// App has Project and ProjectEnv loaded: the subpath a volume gets depends on them.
	App *entity.App
	// Kept are mounts the app already has and keeps as they are. They are not
	// rebuilt, but they take part in the pin conflict check - an unchanged mount
	// pinned to one node conflicts with a new one pinned to another.
	Kept []mount.Mount
	New  []*AppMountReq
}

type BuildAppMountsResp struct {
	// Mounts are Kept followed by the built New mounts.
	Mounts []mount.Mount
}
```

In `hivepaas_app/service/volumeservice/service.go`, add to the interface after `EnsureVolumePermissions`:

```go
	// BuildAppMounts turns requested mounts of cluster-volume settings into the
	// docker mounts an app's service carries: it checks every volume is usable
	// from the app's scope, derives each subpath, rewrites a bind volume into a
	// bind mount, fills in driver config, opens up permissions, and refuses a set
	// of mounts that pins the service to more than one node. It does not write the
	// mounts to the service.
	BuildAppMounts(ctx context.Context, db database.IDB, req *BuildAppMountsReq) (*BuildAppMountsResp, error)
```

- [ ] **Step 3: Add the seams the tests replace**

Replace `New` and `service` in `hivepaas_app/service/volumeservice/volumeserviceimpl/service.go`:

```go
func New(
	dockerManager docker.Manager,
	hpAppService hpappservice.Service,

	settingRepo repository.SettingRepo,
) volumeservice.Service {
	svc := &service{
		dockerManager: dockerManager,
		hpAppService:  hpAppService,

		settingRepo: settingRepo,
	}
	svc.makeSubDirInHost = svc.MakeSubDirInHost
	svc.ensureVolumePermissions = svc.EnsureVolumePermissions
	return svc
}

type service struct {
	dockerManager docker.Manager
	hpAppService  hpappservice.Service

	settingRepo repository.SettingRepo

	// makeSubDirInHost and ensureVolumePermissions reach the host through a
	// throwaway container. They are fields so a test of mount building can stand
	// in for the host without a docker daemon.
	makeSubDirInHost        func(ctx context.Context, baseDirInHost, subpath string, requireBaseDirExist bool) error
	ensureVolumePermissions func(ctx context.Context, volMount *mount.Mount, subpaths ...string) error
}
```

Add `"context"` and `"github.com/moby/moby/api/types/mount"` to its imports.

- [ ] **Step 4: Write the failing tests**

`hivepaas_app/service/volumeservice/volumeserviceimpl/app_mounts_test.go`:

```go
package volumeserviceimpl

import (
	"context"
	"slices"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// appMountsSettingRepo answers the two lookups building mounts makes: the
// requested volumes by id, and every volume the app's scope can see.
type appMountsSettingRepo struct {
	repository.SettingRepo
	volumes []*entity.Setting
}

func (f *appMountsSettingRepo) ListByIDs(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, ids []string, _ bool, _ ...bunex.SelectQueryOption,
) ([]*entity.Setting, error) {
	var out []*entity.Setting
	for _, setting := range f.volumes {
		if slices.Contains(ids, setting.ID) {
			out = append(out, setting)
		}
	}
	return out, nil
}

func (f *appMountsSettingRepo) List(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.volumes, nil, nil
}

type recordedHost struct {
	madeSubDirs []string
	permissions int
}

func newAppMountsTest(volumes ...*entity.Setting) (*service, *recordedHost) {
	host := &recordedHost{}
	svc := &service{settingRepo: &appMountsSettingRepo{volumes: volumes}}
	svc.makeSubDirInHost = func(_ context.Context, baseDir, subpath string, _ bool) error {
		host.madeSubDirs = append(host.madeSubDirs, baseDir+"|"+subpath)
		return nil
	}
	svc.ensureVolumePermissions = func(context.Context, *mount.Mount, ...string) error {
		host.permissions++
		return nil
	}
	return svc, host
}

func mountTestApp() *entity.App {
	return &entity.App{
		ID:         "app-1",
		Key:        "web",
		Project:    &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Key: "prod"},
	}
}

func scopedVolume(
	t *testing.T, id, refID string, scope base.ObjectScopeType, vol *entity.ClusterVolume,
) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, RefID: refID, Type: base.SettingTypeClusterVolume, Scope: scope}
	assert.NoError(t, setting.SetData(vol))
	return setting
}

// A bind volume that stays a TypeVolume mount names a volume the target node has
// never heard of, which is the silent-empty-volume failure bind rewriting exists
// to prevent.
func TestBuildAppMountsRewritesABindVolumeIntoABindMount(t *testing.T) {
	svc, host := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeApp, &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data", "o": "bind"},
	}))

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-1", Target: "/data"}},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Mounts, 1)
	assert.Equal(t, mount.TypeBind, resp.Mounts[0].Type)
	assert.Equal(t, "/srv/data/web", resp.Mounts[0].Source)
	assert.True(t, resp.Mounts[0].BindOptions.CreateMountpoint)
	assert.Equal(t, []string{"/srv/data|web"}, host.madeSubDirs)
}

// Without the driver config in the spec, the node running the task creates an
// empty default volume under the same name and the app writes to local disk.
func TestBuildAppMountsFillsInTheDriverConfigForANonBindVolume(t *testing.T) {
	svc, host := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeProject, &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
	}))

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{
			Type: mount.TypeVolume, Source: "vol-1", Target: "/data",
			VolumeOptions: &volumeservice.AppMountVolumeOptions{Subpath: "files"},
		}},
	})

	assert.NoError(t, err)
	mnt := resp.Mounts[0]
	assert.Equal(t, mount.TypeVolume, mnt.Type)
	assert.Equal(t, "hp-vol-1", mnt.Source)
	assert.Equal(t, "prod/web/files", mnt.VolumeOptions.Subpath, "a project volume is shared, so the app gets env/app")
	assert.Equal(t, "local", mnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/exports/data", mnt.VolumeOptions.DriverConfig.Options["device"])
	assert.Equal(t, 1, host.permissions)
}

func TestBuildAppMountsKeepsTheKeptMountsFirst(t *testing.T) {
	svc, _ := newAppMountsTest(scopedVolume(t, "vol-1", "hp-vol-1", base.ObjectScopeApp, &entity.ClusterVolume{}))
	kept := mount.Mount{Type: mount.TypeVolume, Source: "hp-old", Target: "/old"}

	resp, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App:  mountTestApp(),
		Kept: []mount.Mount{kept},
		New:  []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-1", Target: "/new"}},
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"/old", "/new"}, []string{resp.Mounts[0].Target, resp.Mounts[1].Target})
	assert.Equal(t, kept, resp.Mounts[0], "a kept mount is not rebuilt")
}

func TestBuildAppMountsRefusesAnUnknownVolume(t *testing.T) {
	svc, _ := newAppMountsTest()

	_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-missing", Target: "/data"}},
	})

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestBuildAppMountsRefusesAnotherMountType(t *testing.T) {
	svc, _ := newAppMountsTest()

	_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App: mountTestApp(),
		New: []*volumeservice.AppMountReq{{Type: mount.TypeBind, Source: "/srv", Target: "/data"}},
	})

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

func TestBuildAppMountsRefusesMountsPinnedToTwoNodes(t *testing.T) {
	pgdata := newVolumeSetting(t, "vol-pgdata", "ref-pgdata", "pgdata", "node-1")
	uploads := newVolumeSetting(t, "vol-uploads", "ref-uploads", "uploads", "node-2")
	svc, _ := newAppMountsTest(pgdata, uploads)

	_, err := svc.BuildAppMounts(context.Background(), nil, &volumeservice.BuildAppMountsReq{
		App:  mountTestApp(),
		Kept: []mount.Mount{{Type: mount.TypeVolume, Source: "ref-pgdata", Target: "/pg"}},
		New:  []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: "vol-uploads", Target: "/uploads"}},
	})

	detail := clientVisibleDetail(t, err)
	assert.Contains(t, detail, "pgdata")
	assert.Contains(t, detail, "uploads")
}
```

- [ ] **Step 5: Run them to see them fail**

Run: `go test ./hivepaas_app/service/volumeservice/volumeserviceimpl/`
Expected: FAIL - `svc.BuildAppMounts undefined`. (The moved helper tests compile and pass already.)

- [ ] **Step 6: Write the service method**

`hivepaas_app/service/volumeservice/volumeserviceimpl/app_mounts.go`:

```go
package volumeserviceimpl

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) BuildAppMounts(
	ctx context.Context,
	db database.IDB,
	req *volumeservice.BuildAppMountsReq,
) (*volumeservice.BuildAppMountsResp, error) {
	app := req.App

	volumeIDs := make([]string, 0, len(req.New))
	for _, mnt := range req.New {
		// For custom mounts, only support type Volume and Cluster
		if mnt.Type != mount.TypeVolume && mnt.Type != mount.TypeCluster {
			return nil, hperrors.Wrap(hperrors.ErrUnsupported).
				WithParam("Name", fmt.Sprintf("Mount type '%v'", mnt.Type))
		}
		volumeIDs = append(volumeIDs, mnt.Source)
	}

	// Validate volumes can be used by the project
	volumes, err := s.settingRepo.ListByIDs(ctx, db, app.GetObjectScope(), volumeIDs, true,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	volumeMap := entityutil.SliceToIDMap(volumes)

	mounts := slices.Clone(req.Kept)
	for _, mnt := range req.New {
		setting, found := volumeMap[mnt.Source]
		if !found {
			return nil, hperrors.NewNotFound("Volume").WithMsgLog("volume %v not found", mnt.Source)
		}
		dockerMnt, err := s.buildAppMount(ctx, app, mnt, setting)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		mounts = append(mounts, *dockerMnt)
	}

	// Every volume visible to the app's scope, not just the requested ones: a
	// pinned volume behind an unchanged bind mount carries no id to look it up
	// by, so resolving the mounts back to their pins needs the same candidate set
	// placementserviceimpl matches mounts against.
	scopeVolumes, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// This is the last point before the mounts are saved where a contradiction
	// can be caught: applying a contradictory pin set silently drops the
	// placement constraint instead of failing.
	if err = refuseConflictingVolumePins(mounts, scopeVolumes); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumeservice.BuildAppMountsResp{Mounts: mounts}, nil
}

func (s *service) buildAppMount(
	ctx context.Context,
	app *entity.App,
	mnt *volumeservice.AppMountReq,
	setting *entity.Setting,
) (*mount.Mount, error) {
	vol, err := setting.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dockerMnt := &mount.Mount{
		Type:        mnt.Type,
		Source:      setting.RefID,
		Target:      mnt.Target,
		ReadOnly:    mnt.ReadOnly,
		Consistency: mnt.Consistency,
	}

	s.buildDockerMount(ctx, dockerMnt, mnt, vol, setting, app)

	// Ensure full permissions (0777) on all mounted volume subpaths before starting/updating the service
	if dockerMnt.Type == mount.TypeVolume || dockerMnt.Type == mount.TypeCluster {
		subpath := ""
		if dockerMnt.VolumeOptions != nil {
			subpath = dockerMnt.VolumeOptions.Subpath
		}
		_ = s.ensureVolumePermissions(ctx, dockerMnt, subpath)
	}
	return dockerMnt, nil
}

func (s *service) buildDockerMount(
	ctx context.Context,
	dockerMnt *mount.Mount,
	mnt *volumeservice.AppMountReq,
	vol *entity.ClusterVolume,
	setting *entity.Setting,
	app *entity.App,
) {
	subpath := calcMountSubpath(app, mnt, setting)
	s.useBindMountIfAppropriate(ctx, dockerMnt, vol, subpath)

	switch dockerMnt.Type {
	case mount.TypeVolume:
		if opts := mnt.VolumeOptions; opts != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath:      subpath,
				NoCopy:       opts.NoCopy,
				Labels:       opts.Labels,
				DriverConfig: opts.DriverConfig,
			}
		}
		applyVolumeDriverConfigUnlessOverridden(dockerMnt, vol)
	case mount.TypeCluster:
		if opts := mnt.ClusterOptions; opts != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath:      subpath,
				NoCopy:       opts.NoCopy,
				Labels:       opts.Labels,
				DriverConfig: opts.DriverConfig,
			}
		}
	case mount.TypeBind, mount.TypeImage, mount.TypeTmpfs, mount.TypeNamedPipe:
	}
}

func (s *service) useBindMountIfAppropriate(
	ctx context.Context,
	dockerMnt *mount.Mount,
	vol *entity.ClusterVolume,
	subpath string,
) {
	directory, propagation, ok := bindMountTarget(vol, subpath)
	if !ok {
		return
	}
	if err := s.makeSubDirInHost(ctx, vol.DriverOpts["device"], subpath, true); err != nil {
		return
	}

	dockerMnt.Type = mount.TypeBind
	dockerMnt.Source = directory
	dockerMnt.BindOptions = &mount.BindOptions{
		// Kept for volumes with no pin, which claim every node reaches the same
		// data: creating the directory there is the right thing. A pinned volume
		// is kept on its node by a placement constraint instead.
		CreateMountpoint: true,
	}
	if propagation != "" {
		dockerMnt.BindOptions.Propagation = propagation
	}
	// Reset all other kind of options
	dockerMnt.VolumeOptions = nil
	dockerMnt.ClusterOptions = nil
	dockerMnt.TmpfsOptions = nil
	dockerMnt.ImageOptions = nil
}

func calcMountSubpath(
	app *entity.App,
	mnt *volumeservice.AppMountReq,
	setting *entity.Setting,
) string {
	var subpath string
	switch setting.Scope {
	case base.ObjectScopeGlobal:
		subpath = fmt.Sprintf("%v/%v/%v", app.Project.Key, app.ProjectEnv.Key, app.Key)
	case base.ObjectScopeProject:
		subpath = fmt.Sprintf("%v/%v", app.ProjectEnv.Key, app.Key)
	case base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		subpath = app.Key
	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
	}

	if mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil {
		subpath = filepath.Join(subpath, mnt.VolumeOptions.Subpath)
	}
	if mnt.Type == mount.TypeCluster && mnt.ClusterOptions != nil {
		subpath = filepath.Join(subpath, mnt.ClusterOptions.Subpath)
	}
	return subpath
}
```

- [ ] **Step 7: Point the storage usecase at it**

Replace `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go` with:

```go
package appsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppStorageSettingsReq,
) (*appsettingsdto.UpdateAppStorageSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppStorageSettingsData{}
		err := uc.loadAppStorageSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Building the mounts also refuses a set that pins the service to more
		// than one node, which has to happen before the mounts reach the service.
		built, err := uc.volumeService.BuildAppMounts(ctx, db, &volumeservice.BuildAppMountsReq{
			App:  data.App,
			Kept: data.KeptMounts,
			New:  data.NewMounts,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.FinalMounts = built.Mounts

		err = uc.applyAppStorageSettings(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordAppUpdate(ctx, db, auth, data.App, base.AuditLogSourceAPIUpdate, "storage", auditdetail.New().
			Set("mountCount", len(data.FinalMounts)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppStorageSettingsResp{}, nil
}

type updateAppStorageSettingsData struct {
	App         *entity.App
	Service     *swarm.Service
	KeptMounts  []mount.Mount
	NewMounts   []*volumeservice.AppMountReq
	FinalMounts []mount.Mount
}

func (uc *UC) loadAppStorageSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	// Calculate mount keys of existing mounts to distinguish new changes
	currMounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
	mapCurrMountByKey := make(map[string]*mount.Mount, len(currMounts))
	for i := range currMounts {
		mnt := &currMounts[i]
		mapCurrMountByKey[uc.calcMountKey(mnt)] = mnt
	}

	for _, reqMnt := range req.Mounts {
		if existingMount, exists := mapCurrMountByKey[reqMnt.Key]; reqMnt.Key != "" && exists {
			data.KeptMounts = append(data.KeptMounts, *existingMount) // unchanged mount
			continue
		}
		data.NewMounts = append(data.NewMounts, toAppMountReq(reqMnt))
	}
	return nil
}

// toAppMountReq carries a requested mount to volumeservice, which cannot take
// the DTO.
func toAppMountReq(reqMnt *appsettingsdto.Mount) *volumeservice.AppMountReq {
	out := &volumeservice.AppMountReq{
		Type:        reqMnt.Type,
		Source:      reqMnt.Source,
		Target:      reqMnt.Target,
		ReadOnly:    reqMnt.ReadOnly,
		Consistency: reqMnt.Consistency,
	}
	if opts := reqMnt.VolumeOptions; opts != nil {
		out.VolumeOptions = toAppMountVolumeOptions(opts)
	}
	if opts := reqMnt.ClusterOptions; opts != nil {
		out.ClusterOptions = toAppMountVolumeOptions(&opts.VolumeOptions)
	}
	return out
}

func toAppMountVolumeOptions(opts *appsettingsdto.VolumeOptions) *volumeservice.AppMountVolumeOptions {
	out := &volumeservice.AppMountVolumeOptions{
		Subpath: opts.Subpath,
		NoCopy:  opts.NoCopy,
		Labels:  opts.Labels,
	}
	if driver := opts.DriverConfig; driver != nil {
		out.DriverConfig = &mount.Driver{Name: driver.Name, Options: driver.Options}
	}
	return out
}

// applyAppStorageSettings writes the new mounts and the placement constraints
// they imply in a single service update.
//
// Writing the mounts rolls the service's tasks, so the volume pin has to be in
// the spec this call sends rather than in whatever spec is written next. Left to
// the next deploy, a task rescheduled by this very update can land on a node
// holding none of the data - the failure the pin exists to prevent, reached
// through the branch's own primary flow.
//
// SkipSavingToDocker is what makes one update enough: it has ApplyPlacementSettings
// mutate the spec and stop, instead of saving a second one and rolling the tasks
// again. The call sits inside the callback because a retry re-inspects the
// service and starts over from a spec carrying neither the mounts nor the
// constraint.
func (uc *UC) applyAppStorageSettings(
	ctx context.Context,
	db database.IDB,
	data *updateAppStorageSettingsData,
) error {
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.Service.ID, data.Service,
		func(_ int, service *swarm.Service) (bool, error) {
			service.Spec.TaskTemplate.ContainerSpec.Mounts = data.FinalMounts

			// The pins are resolved from the mounts just written, so this reads the
			// storage settings being saved and not the ones being replaced.
			_, err := uc.placementService.ApplyPlacementSettings(ctx, db,
				&placementservice.ApplyPlacementSettingsReq{
					App:                data.App,
					Service:            service,
					SkipSavingToDocker: true,
				})
			if err != nil {
				return false, hperrors.Wrap(err)
			}
			return true, nil
		}, defaultServiceRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
```

- [ ] **Step 8: Run the moved and the new tests, and the storage usecase's**

Run: `go build ./... && go test ./hivepaas_app/service/volumeservice/... ./hivepaas_app/usecase/appsettingsuc/ ./hivepaas_app/service/placementservice/...`
Expected: PASS. `TestApplyAppStorageSettingsPutsThePinConstraintInTheSameUpdateAsTheMounts` still passes unchanged: it builds `updateAppStorageSettingsData` with `App`, `Service` and `FinalMounts`, all kept.

- [ ] **Step 9: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add -A hivepaas_app/service/volumeservice hivepaas_app/usecase/appsettingsuc
git commit -m "refactor(volume): build app mounts in volumeservice instead of the storage usecase"
```

---

## Task 9: Build an `AppDoc` into an app - `specservice.BuildApp`

**Files:**
- Modify: `hivepaas_app/service/specservice/service.go`
- Modify: `hivepaas_app/service/specservice/types.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/service.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/build.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/build_deployment.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/build_settings.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/export_test.go` (the `New(...)` call)
- Modify: `hivepaas_app/usecase/specuc/export_test.go` (`fakeSpecService`)
- Modify: `hivepaas_app/hperrors/errors_spec.go`, `hivepaas_app/pkg/translation/messages/en/errors.spec.en.toml`
- Test: `hivepaas_app/service/specservice/specserviceimpl/build_test.go`

**Interfaces:**
- Consumes: `specmodel.CheckBuildable`, `PresentBlocks`, `Block*`, `BuildableBlocks` (Task 2); `volumeservice.BuildAppMounts`, `AppMountReq`, `AppMountVolumeOptions`, `BuildAppMountsReq` (Task 8).
- Produces:
  - `type specservice.BuildAppReq struct { App *entity.App; Doc *specmodel.AppDoc; Spec *swarm.ServiceSpec; TimeNow time.Time }`
  - `type specservice.BuildAppResp struct { Settings []*entity.Setting }`
  - `specservice.Service.BuildApp(ctx context.Context, db database.IDB, req *BuildAppReq) (*BuildAppResp, error)` - returns `app-deployment`, `app-kind`, `env-var` and `app-routing` settings for the blocks present, and writes healthcheck, resources and mounts into `req.Spec`
  - `specserviceimpl.New` gains a `volumeService volumeservice.Service` parameter
  - `hperrors.ErrSpecBlockInvalid`

Resources are truncated as `resource_settings_update.go` does. A healthcheck is written in the form docker runs it: a `CMD` command is split into an argv, a `CMD-SHELL` command stays one string. That is deliberately not what `container_settings_update.go` does today - it splits `CMD-SHELL` commands too, and docker then hands the shell only the first word (`sh -c pg_isready -U app` runs `pg_isready` with no arguments). That bug is reported separately rather than copied. The image is not written into the spec: it arrives through the deployment (Task 13), like any image app.

- [ ] **Step 1: Declare the error**

Add to `hivepaas_app/hperrors/errors_spec.go`:

```go
	ErrSpecBlockInvalid            = NewErr(ErrArgumentInvalid, "ERR_SPEC_BLOCK_INVALID")
```

Add to `errors.spec.en.toml`:

```toml
ERR_SPEC_BLOCK_INVALID = "Part of this configuration is invalid"
```

- [ ] **Step 2: Extend the interface and constructor**

Add to `hivepaas_app/service/specservice/types.go` (add imports `time`, `github.com/moby/moby/api/types/swarm`):

```go
// BuildAppReq asks for an AppDoc to be built into an app being provisioned.
type BuildAppReq struct {
	// App is the app being provisioned: ID, Key, Project and ProjectEnv set.
	App *entity.App
	Doc *specmodel.AppDoc
	// Spec is the app's initial service spec. The build writes into it.
	Spec    *swarm.ServiceSpec
	TimeNow time.Time
}

type BuildAppResp struct {
	// Settings replace the app's default settings of the same type.
	Settings []*entity.Setting
}
```

In `hivepaas_app/service/specservice/service.go`, add to `Service`:

```go
	// BuildApp turns an AppDoc into the settings and service spec of an app being
	// provisioned. It persists nothing and creates no service: the caller does.
	// Only what specmodel.CheckBuildable accepts can be built.
	//
	// TODO: spec import - import builds each app of a bundle through this.
	BuildApp(ctx context.Context, db database.IDB, req *BuildAppReq) (*BuildAppResp, error)
```

In `hivepaas_app/service/specservice/specserviceimpl/service.go`, add the dependency:

```go
func New(
	appRepo repository.AppRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,

	clusterService clusterservice.Service,
	volumeService volumeservice.Service,
) specservice.Service {
	svc := &service{
		appRepo:        appRepo,
		projectEnvRepo: projectEnvRepo,
		projectRepo:    projectRepo,
		settingRepo:    settingRepo,

		clusterService: clusterService,
		volumeService:  volumeService,
	}
	svc.loadOwned = svc.loadOwnedFromRepo
	return svc
}
```

and the field `volumeService volumeservice.Service` beside `clusterService` in `service`, importing `github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice`.

In `hivepaas_app/service/specservice/specserviceimpl/export_test.go`, the `New(` call in `exportFixture` gains a last argument `nil,` after the `&fakeClusterService{...}` argument.

In `hivepaas_app/usecase/specuc/export_test.go`, embed the interface in the fake so it keeps compiling as the interface grows:

```go
type fakeSpecService struct {
	specservice.Service
	lastReq *specservice.ExportReq
	err     error
}
```

- [ ] **Step 3: Write the failing tests**

`hivepaas_app/service/specservice/specserviceimpl/build_test.go`:

```go
package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

const buildDocYAML = `
deployment:
  source:
    activeMethod: image
    imageSource: {image: "postgres:17.6-alpine3.22"}
    command: postgres -c max_connections=200
    workingDir: /
  storage:
    mounts:
      /var/lib/postgresql/data: {type: volume, source: vol-1, volumeOptions: {subpath: data}}
  container:
    healthcheck: {enabled: true, mode: CMD-SHELL, command: pg_isready -U app, interval: 10s, retries: 5}
  resources:
    reservations: {cpus: 0.5, memory: 256mb}
    limits: {cpus: 1, memory: 512mb, pids: 100}
settings:
  kind:
    category: database
    engine: postgres
    version: "17"
    database: {dbName: app, username: app, password: s3cret}
  envVars:
    data:
      - {k: POSTGRES_PASSWORD, v: "${HIVEPAAS_PASSWORD}"}
  routing: {port: 5432}
`

type fakeBuildVolumeService struct {
	volumeservice.Service
	req *volumeservice.BuildAppMountsReq
}

func (f *fakeBuildVolumeService) BuildAppMounts(
	_ context.Context, _ database.IDB, req *volumeservice.BuildAppMountsReq,
) (*volumeservice.BuildAppMountsResp, error) {
	f.req = req
	mounts := make([]mount.Mount, 0, len(req.New))
	for _, m := range req.New {
		mounts = append(mounts, mount.Mount{Type: m.Type, Source: "docker-" + m.Source, Target: m.Target})
	}
	return &volumeservice.BuildAppMountsResp{Mounts: mounts}, nil
}

func buildDoc(t *testing.T, text string) *specmodel.AppDoc {
	t.Helper()
	doc := &specmodel.AppDoc{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), doc))
	return doc
}

func buildReq(t *testing.T, text string) *specservice.BuildAppReq {
	t.Helper()
	return &specservice.BuildAppReq{
		App: &entity.App{
			ID: "app-1", Key: "db",
			Project:    &entity.Project{ID: "p1", Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{ID: "p1:dev", Key: "dev", Name: "dev"},
		},
		Doc: buildDoc(t, text),
		Spec: &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Image: "busybox:latest"},
		}},
		TimeNow: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
	}
}

func TestBuildAppBuildsEverySupportedBlock(t *testing.T) {
	useDataKey(t)
	volumes := &fakeBuildVolumeService{}
	svc := &service{volumeService: volumes}
	req := buildReq(t, buildDocYAML)

	resp, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	byType := map[base.SettingType]*entity.Setting{}
	for _, setting := range resp.Settings {
		assert.Equal(t, base.ObjectScopeApp, setting.Scope)
		assert.Equal(t, "app-1", setting.ObjectID)
		assert.Equal(t, base.SettingStatusActive, setting.Status)
		assert.Equal(t, req.TimeNow, setting.CreatedAt)
		byType[setting.Type] = setting
	}
	assert.Len(t, byType, 4)

	deployment := byType[base.SettingTypeAppDeployment].MustAsAppDeploymentSettings()
	assert.Equal(t, base.DeploymentMethodImage, deployment.ActiveMethod)
	assert.Equal(t, "postgres:17.6-alpine3.22", deployment.ImageSource.Image)
	assert.Equal(t, "postgres -c max_connections=200", deployment.Command)

	kind := byType[base.SettingTypeAppKind].MustAsAppKindSettings()
	assert.Equal(t, base.AppCategoryDatabase, kind.Category)
	password, err := kind.Database.Password.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "s3cret", password)

	assert.Equal(t, "${HIVEPAAS_PASSWORD}", byType[base.SettingTypeEnvVar].MustAsEnvVars().Data[0].Value)
	assert.Equal(t, 5432, byType[base.SettingTypeAppRouting].MustAsAppRoutingSettings().Port)

	containerSpec := req.Spec.TaskTemplate.ContainerSpec
	assert.Equal(t, []string{"CMD-SHELL", "pg_isready -U app"}, containerSpec.Healthcheck.Test,
		"a shell command stays one string, as the shell needs it")
	assert.Equal(t, 10*time.Second, containerSpec.Healthcheck.Interval)
	assert.Equal(t, 5, containerSpec.Healthcheck.Retries)

	limits := req.Spec.TaskTemplate.Resources.Limits
	assert.Equal(t, int64(1_000_000_000), limits.NanoCPUs)
	assert.Equal(t, int64(512<<20), limits.MemoryBytes)
	assert.Equal(t, int64(100), limits.Pids)
	assert.Equal(t, int64(256<<20), req.Spec.TaskTemplate.Resources.Reservations.MemoryBytes)

	assert.Equal(t, "docker-vol-1", containerSpec.Mounts[0].Source)
	assert.Equal(t, "/var/lib/postgresql/data", containerSpec.Mounts[0].Target)
	assert.Equal(t, "data", volumes.req.New[0].VolumeOptions.Subpath)
	assert.Equal(t, "busybox:latest", containerSpec.Image, "the image arrives with the deployment")
}

func TestBuildAppAlwaysGivesAVolumeMountOptions(t *testing.T) {
	useDataKey(t)
	volumes := &fakeBuildVolumeService{}
	svc := &service{volumeService: volumes}

	_, err := svc.BuildApp(context.Background(), nil,
		buildReq(t, "deployment:\n  storage:\n    mounts:\n      /data: {type: volume, source: vol-1}\n"))

	assert.NoError(t, err)
	assert.NotNil(t, volumes.req.New[0].VolumeOptions,
		"without options a non-bind volume mount would get no app subpath and share the volume's root")
}

func TestBuildAppSplitsACmdHealthcheck(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, "deployment:\n  container:\n    healthcheck: {enabled: true, mode: CMD, command: \"pg_isready -U app\"}\n")

	_, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	assert.Equal(t, []string{"CMD", "pg_isready", "-U", "app"}, req.Spec.TaskTemplate.ContainerSpec.Healthcheck.Test)
}

// Export maps a service back into an AppDoc; BuildApp maps an AppDoc into a
// service. For the swarm-side blocks the two must agree, or a template and an
// export of the app it created would disagree about the same app.
func TestBuildAppAgreesWithExport(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, buildDocYAML)

	resp, err := svc.BuildApp(context.Background(), nil, req)
	assert.NoError(t, err)

	assert.Equal(t, req.Doc.Deployment.Container.Healthcheck,
		mapHealthcheck(req.Spec.TaskTemplate.ContainerSpec.Healthcheck))
	assert.Equal(t, req.Doc.Deployment.Resources, mapResources(&req.Spec.TaskTemplate))

	for _, setting := range resp.Settings {
		if setting.Type != base.SettingTypeAppDeployment {
			continue
		}
		body, err := renderSetting(setting, newRefIndex(), specmodel.SecretsModeOmit)
		assert.NoError(t, err)
		for key, value := range req.Doc.Deployment.Source {
			assert.Equal(t, value, body[key], key)
		}
	}
}

func TestBuilderRegistryCoversEveryBuildableBlock(t *testing.T) {
	registered := slices.Sorted(maps.Keys((&service{}).builders()))
	assert.Equal(t, slices.Sorted(slices.Values(specmodel.BuildableBlocks)), registered)
}

func TestBuildAppRefuses(t *testing.T) {
	cases := map[string]struct {
		doc  string
		want error
	}{
		"no image": {
			"deployment:\n  source: {activeMethod: image}\n", hperrors.ErrSpecBlockInvalid,
		},
		"reserved env var": {
			"settings:\n  envVars: {data: [{k: HIVEPAAS_PASSWORD, v: x}]}\n", hperrors.ErrSpecBlockInvalid,
		},
		"duplicate env var": {
			"settings:\n  envVars: {data: [{k: A, v: x}, {k: A, v: y}]}\n", hperrors.ErrSpecBlockInvalid,
		},
		"system env var": {
			"settings:\n  envVars: {data: [{k: A, v: x, system: true}]}\n", hperrors.ErrSpecBlockInvalid,
		},
		"unknown kind category": {
			"settings:\n  kind: {category: queue}\n", hperrors.ErrSpecBlockInvalid,
		},
		"unknown kind field": {
			"settings:\n  kind: {category: cache, flavor: x}\n", hperrors.ErrSpecBlockInvalid,
		},
		"port out of range": {
			"settings:\n  routing: {port: 70000}\n", hperrors.ErrSpecBlockInvalid,
		},
		"unbuildable block": {
			"deployment:\n  networks: {dnsConfig: {nameservers: [1.1.1.1]}}\n", hperrors.ErrSpecBlockUnsupported,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			useDataKey(t)
			svc := &service{volumeService: &fakeBuildVolumeService{}}
			_, err := svc.BuildApp(context.Background(), nil, buildReq(t, tc.doc))
			assert.ErrorIs(t, err, tc.want)
		})
	}
}
```

- [ ] **Step 4: Run them to see them fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run Build`
Expected: FAIL - `svc.BuildApp undefined`, `svc.builders undefined`.

- [ ] **Step 5: Write the build entry point**

`hivepaas_app/service/specservice/specserviceimpl/build.go`:

```go
package specserviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type blockBuilder func(ctx context.Context, state *buildState) error

// buildState is what the builders of one BuildApp call share.
type buildState struct {
	db       database.IDB
	req      *specservice.BuildAppReq
	settings []*entity.Setting
}

// builders is the registry of what can be built. It has to cover exactly
// specmodel.BuildableBlocks, which TestBuilderRegistryCoversEveryBuildableBlock
// holds it to.
//
// TODO: app templates phase 3 - more blocks. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func (s *service) builders() map[specmodel.Block]blockBuilder {
	return map[specmodel.Block]blockBuilder{
		specmodel.BlockDeploymentSource:     s.buildSource,
		specmodel.BlockDeploymentStorage:    s.buildStorage,
		specmodel.BlockContainerHealthcheck: s.buildHealthcheck,
		specmodel.BlockDeploymentResources:  s.buildResources,
		specmodel.BlockSettingsKind:         s.buildKind,
		specmodel.BlockSettingsEnvVars:      s.buildEnvVars,
		specmodel.BlockSettingsRouting:      s.buildRouting,
	}
}

func (s *service) BuildApp(
	ctx context.Context,
	db database.IDB,
	req *specservice.BuildAppReq,
) (*specservice.BuildAppResp, error) {
	if err := specmodel.CheckBuildable(req.Doc); err != nil {
		return nil, hperrors.Wrap(err)
	}

	builders := s.builders()
	state := &buildState{db: db, req: req}
	for _, block := range specmodel.PresentBlocks(req.Doc) {
		build, found := builders[block]
		if !found {
			return nil, hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail("%s", block)
		}
		if err := build(ctx, state); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &specservice.BuildAppResp{Settings: state.settings}, nil
}

// addSetting creates a setting row the way the settings usecases create one.
func (state *buildState) addSetting(
	typ base.SettingType,
	version int,
	inheritable bool,
	data entity.SettingData,
) error {
	setting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		ObjectID:    state.req.App.ID,
		Type:        typ,
		Status:      base.SettingStatusActive,
		Inheritable: inheritable,
		Version:     version,
		UpdateVer:   1,
		CreatedAt:   state.req.TimeNow,
		UpdatedAt:   state.req.TimeNow,
	}
	if err := setting.SetData(data); err != nil {
		return hperrors.Wrap(err)
	}
	state.settings = append(state.settings, setting)
	return nil
}

// decodeBlock decodes a settings-shaped block into its entity. Unknown fields
// are refused: a spec field the entity does not have is a field that would not
// be applied.
func decodeBlock(block specmodel.Block, body any, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return invalidBlock(block, "%s", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(out); err != nil {
		return invalidBlock(block, "%s", err.Error())
	}
	return nil
}

func invalidBlock(block specmodel.Block, format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockInvalid).WithExtraDetail("%s: %s", block, fmt.Sprintf(format, args...))
}
```

- [ ] **Step 6: Write the deployment builders**

`hivepaas_app/service/specservice/specserviceimpl/build_deployment.go`:

```go
package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) buildSource(_ context.Context, state *buildState) error {
	block := specmodel.BlockDeploymentSource
	source := &entity.AppDeploymentSettings{}
	if err := decodeBlock(block, state.req.Doc.Deployment.Source, source); err != nil {
		return err
	}
	if source.ActiveMethod != base.DeploymentMethodImage || source.ImageSource == nil ||
		source.ImageSource.Image == "" {
		return invalidBlock(block, "an image to deploy is required")
	}
	return state.addSetting(base.SettingTypeAppDeployment, entity.CurrentAppDeploymentSettingsVersion, true, source)
}

// buildHealthcheck writes a healthcheck in the form docker runs it. CMD is an
// argv, so its command is split; CMD-SHELL is one string handed to the shell,
// so its command is not. A mode left empty means CMD-SHELL, which is what a
// template author writing a shell command expects.
//
// The container settings screen splits a CMD-SHELL command too, and docker then
// hands the shell only its first word: `sh -c pg_isready -U app` runs pg_isready
// with no arguments. That is a bug to fix there, not a behavior to copy here.
func (s *service) buildHealthcheck(_ context.Context, state *buildState) error {
	check := state.req.Doc.Deployment.Container.Healthcheck
	containerSpec := state.req.Spec.TaskTemplate.ContainerSpec
	if !check.Enabled {
		containerSpec.Healthcheck = nil
		return nil
	}

	var test []string
	switch check.Mode {
	case docker.HealthcheckModeCmd:
		command, err := executil.CmdSplit(check.Command)
		if err != nil {
			return invalidBlock(specmodel.BlockContainerHealthcheck, "command: %s", err.Error())
		}
		test = append([]string{string(check.Mode)}, command...)
	case docker.HealthcheckModeInherit, docker.HealthcheckModeCmdShell:
		test = []string{string(docker.HealthcheckModeCmdShell), check.Command}
	case docker.HealthcheckModeNone:
		test = []string{string(docker.HealthcheckModeNone)}
	default:
		return invalidBlock(specmodel.BlockContainerHealthcheck, "mode %q is not one of CMD, CMD-SHELL, NONE",
			check.Mode)
	}

	containerSpec.Healthcheck = &container.HealthConfig{
		Test:          test,
		Interval:      time.Duration(check.Interval),
		Timeout:       time.Duration(check.Timeout),
		StartPeriod:   time.Duration(check.StartPeriod),
		StartInterval: time.Duration(check.StartInterval),
		Retries:       check.Retries,
	}
	return nil
}

// buildResources truncates as the resource settings screen does.
func (s *service) buildResources(_ context.Context, state *buildState) error {
	resources := state.req.Doc.Deployment.Resources
	task := &state.req.Spec.TaskTemplate
	if task.Resources == nil {
		task.Resources = &swarm.ResourceRequirements{}
	}
	if reservations := resources.Reservations; reservations != nil {
		task.Resources.Reservations = &swarm.Resources{
			NanoCPUs:    docker.TruncateCPUsAsNano(reservations.CPUs, docker.MinCPUFraction),
			MemoryBytes: reservations.Memory.Truncate(unit.MB).Bytes(),
		}
	}
	if limits := resources.Limits; limits != nil {
		task.Resources.Limits = &swarm.Limit{
			NanoCPUs:    docker.TruncateCPUsAsNano(limits.CPUs, docker.MinCPUFraction),
			MemoryBytes: limits.Memory.Truncate(unit.MB).Bytes(),
			Pids:        limits.Pids,
		}
	}
	return nil
}

// buildStorage hands the mounts to volumeservice, which builds them the way the
// storage settings screen does.
//
// A mount's source here is a cluster-volume setting id. An export writes docker's
// source there instead - the volume name, or a host path once a bind volume was
// rewritten - which is why storage is left out of the export round-trip test.
// TODO: app templates phase 2 - map one onto the other before merging mounts.
// See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func (s *service) buildStorage(ctx context.Context, state *buildState) error {
	mounts := state.req.Doc.Deployment.Storage.Mounts
	requests := make([]*volumeservice.AppMountReq, 0, len(mounts))
	for _, target := range slices.Sorted(maps.Keys(mounts)) {
		m := mounts[target]
		// Options are always set: volumeservice applies the app's own subpath only
		// to a volume mount that carries them, and a mount without would share the
		// volume's root with every other app on it.
		options := &volumeservice.AppMountVolumeOptions{}
		if m.VolumeOptions != nil {
			options.Subpath = m.VolumeOptions.Subpath
		}
		requests = append(requests, &volumeservice.AppMountReq{
			Type:          m.Type,
			Source:        m.Source,
			Target:        target,
			ReadOnly:      m.ReadOnly,
			VolumeOptions: options,
		})
	}

	built, err := s.volumeService.BuildAppMounts(ctx, state.db, &volumeservice.BuildAppMountsReq{
		App: state.req.App,
		New: requests,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	state.req.Spec.TaskTemplate.ContainerSpec.Mounts = built.Mounts
	return nil
}
```

- [ ] **Step 7: Write the settings builders**

`hivepaas_app/service/specservice/specserviceimpl/build_settings.go`:

```go
package specserviceimpl

import (
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const maxPort = 65535

func (s *service) buildKind(_ context.Context, state *buildState) error {
	block := specmodel.BlockSettingsKind
	kind := &entity.AppKindSettings{}
	body := state.req.Doc.Settings[specmodel.SingletonBlockName(base.SettingTypeAppKind)]
	if err := decodeBlock(block, body, kind); err != nil {
		return err
	}
	if !slices.Contains(base.AllAppCategories, kind.Category) {
		return invalidBlock(block, "category %q is not one of %v", kind.Category, base.AllAppCategories)
	}
	return state.addSetting(base.SettingTypeAppKind, entity.CurrentAppKindSettingsVersion, false, kind)
}

func (s *service) buildEnvVars(_ context.Context, state *buildState) error {
	block := specmodel.BlockSettingsEnvVars
	envVars := &entity.EnvVars{}
	body := state.req.Doc.Settings[specmodel.SingletonBlockName(base.SettingTypeEnvVar)]
	if err := decodeBlock(block, body, envVars); err != nil {
		return err
	}

	seen := map[string]bool{}
	for i, env := range envVars.Data {
		switch {
		case env == nil || env.Key == "":
			return invalidBlock(block, "data[%d] has no key", i)
		case env.IsSystem:
			return invalidBlock(block, "%s: system variables are generated by HivePaaS, not declared", env.Key)
		case !base.IsAppRuntimeEnvAllowed(env.Key):
			return invalidBlock(block, "%s is reserved for HivePaaS", env.Key)
		case seen[env.Key]:
			return invalidBlock(block, "%s is declared twice", env.Key)
		}
		seen[env.Key] = true
	}
	return state.addSetting(base.SettingTypeEnvVar, entity.CurrentEnvVarsVersion, true, envVars)
}

func (s *service) buildRouting(_ context.Context, state *buildState) error {
	block := specmodel.BlockSettingsRouting
	routing := &entity.AppRoutingSettings{}
	body := state.req.Doc.Settings[specmodel.SingletonBlockName(base.SettingTypeAppRouting)]
	if err := decodeBlock(block, body, routing); err != nil {
		return err
	}
	if routing.Port < 0 || routing.Port > maxPort {
		return invalidBlock(block, "port %d is out of range", routing.Port)
	}
	return state.addSetting(base.SettingTypeAppRouting, entity.CurrentAppRoutingSettingsVersion, true, routing)
}
```

- [ ] **Step 8: Run the tests**

Run: `go build ./... && go test ./hivepaas_app/service/specservice/... ./hivepaas_app/usecase/specuc/`
Expected: PASS - the new build tests and every existing export test.

- [ ] **Step 9: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/specservice hivepaas_app/usecase/specuc/export_test.go \
  hivepaas_app/hperrors/errors_spec.go hivepaas_app/pkg/translation/messages/en/errors.spec.en.toml
git commit -m "feat(spec): build an AppDoc into an app's settings and service spec"
```

---
## Task 10: `appprovisionservice` - extracted from `appuc.CreateApp`

**Files:**
- Create: `hivepaas_app/service/appprovisionservice/service.go`
- Create: `hivepaas_app/service/appprovisionservice/types.go`
- Create: `hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/service.go`
- Create: `hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/provision.go`
- Modify: `hivepaas_app/usecase/appuc/create.go`
- Modify: `hivepaas_app/usecase/appuc/uc.go`
- Modify: `hivepaas_app/registry/provides.go`
- Test: `hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/provision_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces:
  - `type appprovisionservice.ConfigureFunc func(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) ([]*entity.Setting, error)`
  - `type appprovisionservice.ProvisionAppReq struct { ProjectID, ProjectEnvID, Name string; Status base.AppStatus; Note string; Tags []string; Configure ConfigureFunc }`
  - `type appprovisionservice.ProvisionAppResp struct { App *entity.App }` - `App.Project`, `App.ProjectEnv`, `App.Settings` and `App.ServiceID` set
  - `appprovisionservice.Service.ProvisionApp(ctx context.Context, db database.IDB, req *ProvisionAppReq) (*ProvisionAppResp, error)`
  - `appprovisionserviceimpl.New(dockerManager docker.Manager, appRepo repository.AppRepo, projectRepo repository.ProjectRepo, appService appservice.Service, clusterService clusterservice.Service, networkService networkservice.Service, placementService placementservice.Service) appprovisionservice.Service`

The code moves from `appuc/create.go` with two deliberate changes, both invisible to `CreateApp`: placement is applied after `Configure` (so a template's mounts are there when volume pins are resolved), and a failure after the swarm service exists removes it inside `ProvisionApp`. `CreateApp` has no tests today, so this task adds them at the new home.

- [ ] **Step 1: Declare the service**

`hivepaas_app/service/appprovisionservice/service.go`:

```go
// Package appprovisionservice creates apps together with their configuration.
// An empty app is the smallest case; an app from a template is the full one.
package appprovisionservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// ProvisionApp creates an app, its swarm service and its settings inside the
	// caller's transaction.
	//
	// The swarm service cannot take part in that transaction. A failure after it
	// is created removes it before ProvisionApp returns; a failure the caller hits
	// afterwards is the caller's to clean up, through the returned app's ServiceID.
	//
	// TODO: spec import - import provisions each app of a bundle through this.
	ProvisionApp(ctx context.Context, db database.IDB, req *ProvisionAppReq) (*ProvisionAppResp, error)
}
```

`hivepaas_app/service/appprovisionservice/types.go`:

```go
package appprovisionservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// ConfigureFunc fills in an app's configuration. It runs once the app and its
// initial service spec are prepared and before the service is created. The
// settings it returns replace the app's default settings of the same type; what
// it writes into spec is created with the service.
type ConfigureFunc func(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) (
	[]*entity.Setting, error)

type ProvisionAppReq struct {
	ProjectID    string
	ProjectEnvID string
	Name         string
	Status       base.AppStatus
	Note         string
	Tags         []string
	// Configure is nil for an empty app.
	Configure ConfigureFunc
}

type ProvisionAppResp struct {
	// App has Project, ProjectEnv and Settings set, and the ServiceID of the
	// service created for it.
	App *entity.App
}
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/provision_test.go`:

```go
package appprovisionserviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var errTestPersist = errors.New("database unavailable")

type fakeProjectRepo struct {
	repository.ProjectRepo
	project *entity.Project
}

func (f *fakeProjectRepo) GetByID(
	_ context.Context, _ database.IDB, _ string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	return f.project, nil
}

type fakeAppRepo struct {
	repository.AppRepo
}

func (f *fakeAppRepo) GetByGlobalKey(
	_ context.Context, _ database.IDB, _, _ string, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	return nil, hperrors.NewNotFound("App")
}

type fakeNetworkService struct {
	networkservice.Service
}

func (f *fakeNetworkService) GetOrCreateProjectNetwork(
	_ context.Context, _ database.IDB, _ *entity.Project, _ string,
) (*entity.Setting, *network.Inspect, error) {
	return nil, nil, nil
}

func (f *fakeNetworkService) GetProjectNetworkName(_ *entity.Project, env string) string {
	return "net-" + env
}

// fakePlacementService records the mounts it saw, to show placement runs after
// Configure has written them.
type fakePlacementService struct {
	placementservice.Service
	sawMounts int
}

func (f *fakePlacementService) ApplyPlacementSettings(
	_ context.Context, _ database.IDB, req *placementservice.ApplyPlacementSettingsReq,
) (*placementservice.ApplyPlacementSettingsResp, error) {
	f.sawMounts = len(req.Service.Spec.TaskTemplate.ContainerSpec.Mounts)
	return &placementservice.ApplyPlacementSettingsResp{Service: req.Service}, nil
}

type fakeDockerManager struct {
	docker.Manager
	created *swarm.ServiceSpec
}

func (f *fakeDockerManager) ServiceCreate(
	_ context.Context, spec *swarm.ServiceSpec, _ ...docker.ServiceCreateOption,
) (*client.ServiceCreateResult, error) {
	f.created = spec
	return &client.ServiceCreateResult{ID: "svc-1"}, nil
}

type fakeClusterService struct {
	clusterservice.Service
	removed []string
}

func (f *fakeClusterService) ServiceRemove(_ context.Context, serviceID string, _ int, _ time.Duration) error {
	f.removed = append(f.removed, serviceID)
	return nil
}

type fakeAppService struct {
	appservice.Service
	persisted *appservice.PersistingAppData
	err       error
}

func (f *fakeAppService) PersistAppData(
	_ context.Context, _ database.IDB, data *appservice.PersistingAppData,
) error {
	f.persisted = data
	return f.err
}

type provisionFakes struct {
	docker    *fakeDockerManager
	cluster   *fakeClusterService
	apps      *fakeAppService
	placement *fakePlacementService
}

func newProvisionTest(t *testing.T) (*service, *provisionFakes) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: config.EnvProd})
	t.Cleanup(func() { config.SetCurrent(previous) })

	env := &entity.ProjectEnv{ID: "p1:prod", Key: "prod", Name: "prod", Status: base.ProjectStatusActive}
	project := &entity.Project{ID: "p1", Key: "shop", Name: "Shop", Status: base.ProjectStatusActive,
		ProjectEnvs: []*entity.ProjectEnv{env}}

	fakes := &provisionFakes{
		docker:    &fakeDockerManager{},
		cluster:   &fakeClusterService{},
		apps:      &fakeAppService{},
		placement: &fakePlacementService{},
	}
	svc := &service{
		dockerManager:    fakes.docker,
		appRepo:          &fakeAppRepo{},
		projectRepo:      &fakeProjectRepo{project: project},
		appService:       fakes.apps,
		clusterService:   fakes.cluster,
		networkService:   &fakeNetworkService{},
		placementService: fakes.placement,
	}
	return svc, fakes
}

func settingTypes(settings []*entity.Setting) []base.SettingType {
	types := make([]base.SettingType, 0, len(settings))
	for _, setting := range settings {
		types = append(types, setting.Type)
	}
	return types
}

func TestProvisionAppCreatesAnEmptyApp(t *testing.T) {
	svc, fakes := newProvisionTest(t)

	resp, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "Web API", Status: base.AppStatusActive,
		Tags: []string{"api"},
	})

	assert.NoError(t, err)
	app := resp.App
	assert.Equal(t, "svc-1", app.ServiceID)
	assert.Equal(t, "shop", app.Project.Key)
	assert.Equal(t, "prod", app.ProjectEnv.Key)
	assert.Equal(t, []base.SettingType{base.SettingTypeAppRouting, base.SettingTypeAppFeatures},
		settingTypes(app.Settings))

	assert.Equal(t, dockerImageInit, fakes.docker.created.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, app.GlobalKey, fakes.docker.created.Name)
	assert.Equal(t, "net-prod", fakes.docker.created.TaskTemplate.Networks[0].Target)

	assert.Equal(t, []*entity.App{app}, fakes.apps.persisted.UpsertingApps)
	assert.Equal(t, app.Settings, fakes.apps.persisted.UpsertingSettings)
	assert.Equal(t, "api", fakes.apps.persisted.UpsertingTags[0].Tag)
	assert.Empty(t, fakes.cluster.removed)
}

func TestProvisionAppAppliesTheConfiguration(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	routing := &entity.Setting{ID: "set-routing", Type: base.SettingTypeAppRouting}
	kind := &entity.Setting{ID: "set-kind", Type: base.SettingTypeAppKind}

	resp, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
		Configure: func(_ context.Context, _ database.IDB, app *entity.App, spec *swarm.ServiceSpec) (
			[]*entity.Setting, error) {
			assert.NotEmpty(t, app.ID, "the app has its id before it is configured")
			spec.TaskTemplate.ContainerSpec.Mounts = append(spec.TaskTemplate.ContainerSpec.Mounts,
				mountAt("/data"))
			return []*entity.Setting{routing, kind}, nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, []base.SettingType{base.SettingTypeAppFeatures, base.SettingTypeAppRouting, base.SettingTypeAppKind},
		settingTypes(resp.App.Settings), "a configured setting replaces the default of its type")
	assert.Same(t, routing, resp.App.Settings[1])
	assert.Equal(t, 1, fakes.placement.sawMounts, "placement runs after Configure wrote the mounts")
	assert.Len(t, fakes.docker.created.TaskTemplate.ContainerSpec.Mounts, 1)
}

func TestProvisionAppCreatesNoServiceWhenConfigureFails(t *testing.T) {
	svc, fakes := newProvisionTest(t)

	_, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
		Configure: func(context.Context, database.IDB, *entity.App, *swarm.ServiceSpec) ([]*entity.Setting, error) {
			return nil, hperrors.Wrap(hperrors.ErrSpecBlockInvalid)
		},
	})

	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
	assert.Nil(t, fakes.docker.created)
}

func TestProvisionAppRemovesTheServiceWhenPersistingFails(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	fakes.apps.err = errTestPersist

	_, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
	})

	assert.ErrorIs(t, err, errTestPersist)
	assert.Equal(t, []string{"svc-1"}, fakes.cluster.removed)
}

func TestProvisionAppRefusesAnInactiveEnv(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	svc.projectRepo.(*fakeProjectRepo).project.ProjectEnvs[0].Status = base.ProjectStatusDisabled

	_, err := svc.ProvisionApp(context.Background(), nil, &appprovisionservice.ProvisionAppReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "db", Status: base.AppStatusActive,
	})

	assert.ErrorIs(t, err, hperrors.ErrProjectEnvInactive)
	assert.Nil(t, fakes.docker.created)
}

func mountAt(target string) mount.Mount {
	return mount.Mount{Type: mount.TypeVolume, Source: "vol", Target: target}
}
```


- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/appprovisionservice/...`
Expected: FAIL - `undefined: service`, `undefined: dockerImageInit`.

- [ ] **Step 4: Write the implementation**

`hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/service.go`:

```go
package appprovisionserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	projectRepo repository.ProjectRepo,

	appService appservice.Service,
	clusterService clusterservice.Service,
	networkService networkservice.Service,
	placementService placementservice.Service,
) appprovisionservice.Service {
	return &service{
		dockerManager: dockerManager,

		appRepo:     appRepo,
		projectRepo: projectRepo,

		appService:       appService,
		clusterService:   clusterService,
		networkService:   networkService,
		placementService: placementService,
	}
}

type service struct {
	dockerManager docker.Manager

	appRepo     repository.AppRepo
	projectRepo repository.ProjectRepo

	appService       appservice.Service
	clusterService   clusterservice.Service
	networkService   networkservice.Service
	placementService placementservice.Service
}
```

`hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/provision.go`:

```go
package appprovisionserviceimpl

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/apphelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

const (
	dockerImageInit    = "busybox:latest"
	dockerImageInitDev = "crccheck/hello-world:latest"
)

func (s *service) ProvisionApp(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
) (resp *appprovisionservice.ProvisionAppResp, err error) {
	project, projectEnv, err := s.loadProjectEnv(ctx, db, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	app, err := s.newApp(ctx, db, req, project, projectEnv, timeNow)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	svc := initialService(app, s.networkService.GetProjectNetworkName(project, projectEnv.Name))
	settings := defaultAppSettings(app, timeNow)
	if req.Configure != nil {
		configured, configureErr := req.Configure(ctx, db, app, &svc.Spec)
		if configureErr != nil {
			return nil, hperrors.Wrap(configureErr)
		}
		settings = replaceSettingsByType(settings, configured)
	}
	app.Settings = settings

	// After Configure, so volume pins are resolved from the mounts it wrote.
	_, err = s.placementService.ApplyPlacementSettings(ctx, db, &placementservice.ApplyPlacementSettingsReq{
		App:                app,
		Service:            svc,
		SkipSavingToDocker: true,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	created, err := s.dockerManager.ServiceCreate(ctx, &svc.Spec)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if created.ID == "" { // should never happen
		return nil, hperrors.Wrap(hperrors.ErrInfraInternal).WithParam("Error", "empty service ID returned")
	}
	app.ServiceID = created.ID

	defer func() {
		if err != nil {
			_ = s.clusterService.ServiceRemove(context.WithoutCancel(ctx), app.ServiceID,
				clusterservice.ItemRemovalRetryMax, 0)
		}
	}()

	err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingApps:     []*entity.App{app},
		UpsertingTags:     appTags(app, req.Tags),
		UpsertingSettings: settings,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appprovisionservice.ProvisionAppResp{App: app}, nil
}

func (s *service) loadProjectEnv(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
) (*entity.Project, *entity.ProjectEnv, error) {
	project, err := s.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectRelation("ProjectEnvs",
			bunex.SelectWhere("project_env.id = ?", req.ProjectEnvID),
		),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if project.Status != base.ProjectStatusActive {
		return nil, nil, hperrors.Wrap(hperrors.ErrProjectInactive).WithParam("Name", project.Name)
	}
	if len(project.ProjectEnvs) == 0 {
		return nil, nil, hperrors.Wrap(hperrors.ErrProjectEnvNotFound).WithParam("Name", req.ProjectEnvID)
	}
	projectEnv := project.ProjectEnvs[0]
	if projectEnv.Status != base.ProjectStatusActive {
		return nil, nil, hperrors.Wrap(hperrors.ErrProjectEnvInactive).WithParam("Project", project.Name).
			WithParam("Env", projectEnv.Name)
	}
	return project, projectEnv, nil
}

func (s *service) newApp(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ProvisionAppReq,
	project *entity.Project,
	projectEnv *entity.ProjectEnv,
	timeNow time.Time,
) (*entity.App, error) {
	app := &entity.App{
		ID:           gofn.Must(ulid.NewStringULID()),
		ProjectID:    project.ID,
		Project:      project,
		ProjectEnvID: projectEnv.ID,
		ProjectEnv:   projectEnv,
		Key:          projecthelper.CalcAppKey(req.Name),
		Name:         req.Name,
		Status:       req.Status,
		Note:         req.Note,
		CreatedAt:    timeNow,
		UpdatedAt:    timeNow,
	}
	app.GlobalKey = projecthelper.CalcAppGlobalKey(project.Key, app.Key, projectEnv.Key)

	// App keys must be unique globally
	conflictApp, err := s.appRepo.GetByGlobalKey(ctx, db, "", app.GlobalKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if conflictApp != nil {
		return nil, hperrors.NewAlreadyExist("App").
			WithMsgLog("app unique key '%s' already exists", app.GlobalKey)
	}

	// Create local network for the app to attach
	if _, _, err = s.networkService.GetOrCreateProjectNetwork(ctx, db, project, projectEnv.Key); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return app, nil
}

// initialService is the placeholder service an app starts with. Its image is
// replaced by the app's first deployment.
func initialService(app *entity.App, networkName string) *swarm.Service {
	isDevEnv := config.Current().IsDevEnv()
	appInfo := &apphelper.AppInfo{
		Name: app.Name,
		Key:  app.Key,
		Env:  app.ProjectEnv.Name,
	}
	return &swarm.Service{
		Spec: swarm.ServiceSpec{
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: new(uint64(1)),
				},
			},
			Annotations: swarm.Annotations{
				Name: app.GlobalKey,
				Labels: map[string]string{
					appservice.LabelAppNamespace: app.Project.Key,
					appservice.LabelAppInfo:      apphelper.CalcAppInfoLabel(appInfo),
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image:    gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit),
					Command:  gofn.If(isDevEnv, nil, []string{"sleep", "infinity"}),
					Hostname: app.Key,
					Init:     new(true), // default to use `tini`
					// The app's identity, for the log collector. Container labels
					// rather than service ones: swarm does not pass service labels
					// down, so only these can reach a log line's attrs.
					Labels: appservice.WithAppLogLabels(nil, app),
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{
						Target:  networkName,
						Aliases: []string{app.Key},
					},
				},
				// See DefaultLogDriver for why this is not `local`.
				LogDriver: appservice.DefaultLogDriver(),
			},
		},
	}
}

func defaultAppSettings(app *entity.App, timeNow time.Time) []*entity.Setting {
	// Init empty routing settings
	routingSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		Type:        base.SettingTypeAppRouting,
		Status:      base.SettingStatusActive,
		ObjectID:    app.ID,
		Inheritable: true,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	routingSetting.MustSetData(&entity.AppRoutingSettings{})

	// Init feature settings
	featureSettings := &entity.AppFeatureSettings{}
	entity.InitAppFeatureSettingsDefault(featureSettings)
	featureSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		Type:        base.SettingTypeAppFeatures,
		Status:      base.SettingStatusActive,
		ObjectID:    app.ID,
		Inheritable: true,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	featureSetting.MustSetData(featureSettings)

	return []*entity.Setting{routingSetting, featureSetting}
}

// replaceSettingsByType keeps each default no configured setting has the type
// of, then appends the configured settings.
func replaceSettingsByType(defaults, configured []*entity.Setting) []*entity.Setting {
	out := make([]*entity.Setting, 0, len(defaults)+len(configured))
	for _, setting := range defaults {
		if !slices.ContainsFunc(configured, func(c *entity.Setting) bool { return c.Type == setting.Type }) {
			out = append(out, setting)
		}
	}
	return append(out, configured...)
}

func appTags(app *entity.App, tags []string) []*entity.Tag {
	out := make([]*entity.Tag, 0, len(tags))
	for index, tag := range tags {
		out = append(out, &entity.Tag{ObjectID: app.ID, Tag: tag, Index: index})
	}
	return out
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/appprovisionservice/...`
Expected: PASS.

- [ ] **Step 6: Make `CreateApp` use it**

Replace everything in `hivepaas_app/usecase/appuc/create.go` from the `import` block down to (and including) `func (uc *UC) preparePersistingApp(...)` and `preparePersistingAppService` and `preparePersistingAppSettingsDefault`, keeping `persistingAppData`, `preparePersistingAppBase`, `preparePersistingAppTags` and `persistData` - `update.go` uses them. The file becomes:

```go
package appuc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) CreateApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CreateAppReq,
) (resp *appdto.CreateAppResp, err error) {
	resp = &appdto.CreateAppResp{}
	var createdApp *entity.App

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.ErrPanic)
		}
		if err != nil && createdApp != nil && createdApp.ServiceID != "" {
			_ = uc.clusterService.ServiceRemove(ctx, createdApp.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		provisioned, err := uc.appProvisionService.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
			ProjectID:    req.ProjectID,
			ProjectEnvID: req.ProjectEnvID,
			Name:         req.Name,
			Status:       req.Status,
			Note:         req.Note,
			Tags:         req.Tags,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		createdApp = provisioned.App
		resp.Data = &basedto.ObjectIDResp{ID: createdApp.ID}

		// After persisting and still inside the transaction. ProvisionApp removes
		// the service itself when it fails; a record that fails here errors the
		// call and the deferred cleanup above removes it - so there is no path
		// that leaves a live app behind with nothing recorded.
		return uc.recordAppWrite(ctx, db, auth, createdApp,
			base.AuditLogTypeAppCreate, base.AuditLogSourceAPICreate, "create", auditdetail.New().
				Set("projectId", createdApp.ProjectID).
				Set("envId", createdApp.ProjectEnvID).
				Set("serviceId", createdApp.ServiceID))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

type persistingAppData struct {
	appservice.PersistingAppData
}
```

followed by the unchanged `preparePersistingAppBase`, `preparePersistingAppTags` and `persistData`.

In `hivepaas_app/usecase/appuc/uc.go`: remove the now-unused `projectRepo`, `networkService` and `placementService` fields, parameters, assignments and imports, and add `appProvisionService appprovisionservice.Service` as a field, a `New` parameter after `appCloneService`, and an assignment.

In `hivepaas_app/registry/provides.go`, add `appprovisionserviceimpl.New,` to the service constructors after `appdeploymentserviceimpl.New,`, with its import.

- [ ] **Step 7: Build and run everything app-related**

Run: `go build ./... && go test ./hivepaas_app/usecase/appuc/ ./hivepaas_app/service/appprovisionservice/... ./hivepaas_app/registry/...`
Expected: PASS. If the registry has a wiring test or the build fails in `registry`, the new constructor is not listed or its import is missing.

- [ ] **Step 8: Check the live flow is unchanged**

Run a second backend on its own port, so an instance already serving port 10000 (an IDE run, say) is left alone, and sign in through the dev helper as `docs/DEVELOPMENT.md` §6 does:

```bash
HP_HTTP_SERVER_PORT=10077 HP_RUN_MODE=app make local-app-run &

USER_ID=$(PGPASSWORD=abc123 psql -h localhost -p 35432 -U hivepaas -d hivepaas \
  -tAc "SELECT id FROM users WHERE deleted_at IS NULL ORDER BY created_at LIMIT 1")
TOKEN=$(curl -sS -u hivepaas:abc123 -X POST \
  "http://localhost:10077/_/internal/dev-helper/dev-mode-login?userId=$USER_ID" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["accessToken"])')

curl -sS -X POST "http://localhost:10077/_/projects/<projectID>/<env>/apps" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"provision-check","status":"active","tags":[]}'
```

Expected: `201` with `data.id`; the app shows its routing and feature settings in the dashboard, and `docker service ls` lists its service. Delete the app and stop the second backend afterwards.

- [ ] **Step 9: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/appprovisionservice hivepaas_app/usecase/appuc hivepaas_app/registry/provides.go
git commit -m "refactor(app): provision apps through appprovisionservice"
```

---
## Task 11: Fetching and verifying templates - `apptemplateservice`

**Files:**
- Create: `hivepaas_app/config/app_templates.go`
- Modify: `hivepaas_app/config/config.go`
- Create: `hivepaas_app/service/apptemplateservice/service.go`
- Create: `hivepaas_app/service/apptemplateservice/source.go`
- Create: `hivepaas_app/service/apptemplateservice/types.go`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/source_official.go`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/source_localdir.go`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/categories.yaml`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/tags.yaml`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/templates/demo.yaml`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/icons/demo.svg`
- Modify: `hivepaas_app/hperrors/errors_app_template.go`, `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`
- Modify: `hivepaas_app/registry/provides.go`
- Test: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/source_official_test.go`
- Test: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service_test.go`

**Interfaces:**
- Consumes: `templatemodel` (Task 1), `templaterender.Render`, `Request`, `Result` (Task 5), `templaterepo.Load`, `BuildIndex`, `FindTemplate`, `IndexFile`, `Max*Size` (Task 6), `base.TemplatesRef` and `hpappservice.Service.GetAppReleaseInfo` (existing).
- Produces:
  - `config.AppTemplates struct { Dir string }` at `Config.AppTemplates`, env `HP_TEMPLATES_DIR`
  - `apptemplateservice.Source` interface: `ID() string`, `Revision(ctx) (string, error)`, `Index(ctx) (*templatemodel.Index, error)`, `TemplateFile(ctx, entry *templatemodel.IndexEntry) ([]byte, error)`, `Icon(ctx, sha256 string) ([]byte, error)`
  - `apptemplateservice.Service` interface: `Index(ctx) (*IndexResp, error)`, `Template(ctx, name string) (*TemplateResp, error)`, `Icon(ctx, sha256 string) (*IconResp, error)`, `Render(ctx, req *RenderReq) (*RenderResp, error)`
  - `type IndexResp struct { Source, Revision string; Index *templatemodel.Index }`
  - `type TemplateResp struct { Source, Revision string; Entry *templatemodel.IndexEntry; Template *templatemodel.Template }`
  - `type IconResp struct { Content []byte; ContentType string }`
  - `type RenderReq struct { Name, Version, Variant string; Params map[string]any }`
  - `type RenderResp struct { TemplateResp; Result *templaterender.Result }`
  - `apptemplateserviceimpl.New(hpAppService hpappservice.Service, logger logging.Logger) apptemplateservice.Service`
  - `hperrors.ErrAppTemplatesUnavailable`, `ErrAppTemplateNotFound`, `ErrAppTemplateVerificationFailed`, `ErrAppTemplateIncompatible`

- [ ] **Step 1: Declare the errors and the config**

Add to `hivepaas_app/hperrors/errors_app_template.go`:

```go
	ErrAppTemplatesUnavailable       = NewErr(ErrUnavailable, "ERR_APP_TEMPLATES_UNAVAILABLE")
	ErrAppTemplateNotFound           = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_NOT_FOUND")
	ErrAppTemplateVerificationFailed = NewErr(ErrPreconditionFailed, "ERR_APP_TEMPLATE_VERIFICATION_FAILED")
	ErrAppTemplateIncompatible       = NewErr(ErrNotAllowed, "ERR_APP_TEMPLATE_INCOMPATIBLE")
```

Add to `errors.app_template.en.toml`:

```toml
ERR_APP_TEMPLATES_UNAVAILABLE = "App templates are not available right now"
ERR_APP_TEMPLATE_NOT_FOUND = "App template '{{.Name}}' does not exist"
ERR_APP_TEMPLATE_VERIFICATION_FAILED = "An app template file failed verification and was not used"
ERR_APP_TEMPLATE_INCOMPATIBLE = "App template '{{.Name}}' requires a newer version of HivePaaS"
```

`hivepaas_app/config/app_templates.go`:

```go
package config

// AppTemplates configures where app templates come from.
type AppTemplates struct {
	// Dir reads templates from a checkout of the app-templates repository instead
	// of the signed official source, with no signature and no hashes. It is for
	// authoring templates, and it is honored only in the development environment.
	Dir string `toml:"dir" env:"HP_TEMPLATES_DIR"`
}
```

In `hivepaas_app/config/config.go`, add to `Config` after `Agent`:

```go
	AppTemplates AppTemplates `toml:"app_templates"`
```

- [ ] **Step 2: Declare the service**

`hivepaas_app/service/apptemplateservice/source.go`:

```go
// Package apptemplateservice serves app templates: the catalogue, one template,
// its icon, and a template rendered for creating an app.
package apptemplateservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// Source is where templates come from. Everything that reads templates - the
// store, provisioning, and phase 2's updates - reads them through a Source.
//
// TODO: app templates phase 3 - a git repository an administrator adds, and an
// uploaded file, as further sources. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type Source interface {
	ID() string
	// Revision identifies what is being served, such as a commit.
	Revision(ctx context.Context) (string, error)
	Index(ctx context.Context) (*templatemodel.Index, error)
	// TemplateFile returns an entry's template file, verified wherever the source
	// is able to verify it.
	TemplateFile(ctx context.Context, entry *templatemodel.IndexEntry) ([]byte, error)
	// Icon returns the icon with this sha256. Only an icon the index lists is served.
	Icon(ctx context.Context, sha256 string) ([]byte, error)
}
```

`hivepaas_app/service/apptemplateservice/service.go`:

```go
package apptemplateservice

import "context"

type Service interface {
	Index(ctx context.Context) (*IndexResp, error)
	Template(ctx context.Context, name string) (*TemplateResp, error)
	Icon(ctx context.Context, sha256 string) (*IconResp, error)
	// Render loads a template and renders it for creating an app. A template that
	// needs a newer HivePaaS is refused, and so is a deprecated version.
	Render(ctx context.Context, req *RenderReq) (*RenderResp, error)
}
```

`hivepaas_app/service/apptemplateservice/types.go`:

```go
package apptemplateservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

type IndexResp struct {
	Source   string
	Revision string
	Index    *templatemodel.Index
}

type TemplateResp struct {
	Source   string
	Revision string
	Entry    *templatemodel.IndexEntry
	Template *templatemodel.Template
}

type IconResp struct {
	Content     []byte
	ContentType string
}

type RenderReq struct {
	Name    string
	Version string
	Variant string
	Params  map[string]any
}

type RenderResp struct {
	TemplateResp
	Result *templaterender.Result
}
```

- [ ] **Step 3: Add the test repository**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/categories.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: TemplateCategories
categories:
  - id: databases
    title: Databases
    children:
      - {id: sql, title: SQL}
```

`.../testdata/repo/tags.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: TemplateTags
tags:
  - {id: sql, title: SQL}
```

`.../testdata/repo/templates/demo.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: demo
  title: Demo
  tagline: A demo database
  description: Demo.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
parameters:
  - {name: password, title: Password, type: secret, generate: {length: 16}}
  - {name: dataVolume, title: Data volume, type: volume}
versions:
  - {name: "2", release: "2.1", default: true, image: "demo:2.1.0"}
  - {name: "1", release: "1.9", deprecated: true, image: "demo:1.9.3"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /data: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    kind: {category: database, engine: demo, database: {password: "${{ params.password }}"}}
```

`.../testdata/repo/icons/demo.svg`:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"/>
```

- [ ] **Step 4: Write the failing source tests**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/source_official_test.go`:

```go
package apptemplateserviceimpl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	testRepo       = "hivepaas/app-templates"
	testCommit     = "0123456789abcdef0123456789abcdef01234567"
	testBetaCommit = "89abcdef0123456789abcdef0123456789abcdef"
	testRepoDir    = "testdata/repo"
)

type fakeReleaseInfo struct {
	hpappservice.Service
	info *hpappservice.AppReleaseInfo
}

func (f *fakeReleaseInfo) GetAppReleaseInfo(context.Context) (*hpappservice.AppReleaseInfo, error) {
	return f.info, nil
}

// templatesServer stands in for raw.githubusercontent.com.
type templatesServer struct {
	mu     sync.Mutex
	files  map[string][]byte
	hits   map[string]int
	server *httptest.Server
}

func (ts *templatesServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.hits[r.URL.Path]++
	data, ok := ts.files[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}

func (ts *templatesServer) set(path string, data []byte) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.files["/"+testRepo+"/"+testCommit+"/"+path] = data
}

func (ts *templatesServer) hitCount(path string) int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.hits["/"+testRepo+"/"+testCommit+"/"+path]
}

type officialFixture struct {
	source    *officialSource
	server    *templatesServer
	cacheDir  string
	indexData []byte
	template  []byte
	icon      []byte
	release   *hpappservice.AppReleaseInfo
}

func useEnv(t *testing.T, env string) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: env})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func pinOf(commit string, indexData []byte) *hpappservice.ReleaseInfo {
	sum := sha256.Sum256(indexData)
	return &hpappservice.ReleaseInfo{ReleaseInfo: base.ReleaseInfo{
		AppVersion: "v0.1.1",
		Templates:  &base.TemplatesRef{Repo: testRepo, Commit: commit, IndexSHA256: hex.EncodeToString(sum[:])},
	}}
}

func newOfficialFixture(t *testing.T) *officialFixture {
	t.Helper()
	useEnv(t, config.EnvProd)

	repo, problems, err := templaterepo.Load(os.DirFS(testRepoDir))
	assert.NoError(t, err)
	assert.Empty(t, problems)
	index, err := templaterepo.BuildIndex(repo)
	assert.NoError(t, err)
	indexData, err := templaterepo.MarshalIndex(index)
	assert.NoError(t, err)

	fx := &officialFixture{
		server:    &templatesServer{files: map[string][]byte{}, hits: map[string]int{}},
		cacheDir:  t.TempDir(),
		indexData: indexData,
		template:  repo.FindTemplate("demo").Content,
		icon:      repo.Icons["icons/demo.svg"].Content,
		release:   &hpappservice.AppReleaseInfo{Stable: pinOf(testCommit, indexData)},
	}
	fx.server.server = httptest.NewServer(fx.server)
	t.Cleanup(fx.server.server.Close)
	fx.server.set(templaterepo.IndexFile, indexData)
	fx.server.set("templates/demo.yaml", fx.template)
	fx.server.set("icons/demo.svg", fx.icon)
	fx.source = fx.newSource()
	return fx
}

// newSource is a fresh source - no in-memory index - over the same server and
// disk cache, the way a restarted process sees them.
func (fx *officialFixture) newSource() *officialSource {
	source := newOfficialSource(&fakeReleaseInfo{info: fx.release})
	source.baseURL = fx.server.server.URL
	source.cacheDir = func() string { return fx.cacheDir }
	return source
}

func TestOfficialSourceServesVerifiedFiles(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()

	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	entry := index.FindTemplate("demo")

	template, err := fx.source.TemplateFile(ctx, entry)
	assert.NoError(t, err)
	assert.Equal(t, fx.template, template)

	icon, err := fx.source.Icon(ctx, entry.Icon.SHA256)
	assert.NoError(t, err)
	assert.Equal(t, fx.icon, icon)

	revision, err := fx.source.Revision(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testCommit, revision)

	cached, err := os.ReadFile(filepath.Join(fx.cacheDir, entry.File.SHA256))
	assert.NoError(t, err)
	assert.Equal(t, fx.template, cached, "a verified file is cached under its hash")
}

func TestOfficialSourceServesTheCacheWhenGitHubIsDown(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()
	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	_, err = fx.source.TemplateFile(ctx, index.FindTemplate("demo"))
	assert.NoError(t, err)

	fx.server.server.Close()
	restarted := fx.newSource()

	index, err = restarted.Index(ctx)
	assert.NoError(t, err)
	template, err := restarted.TemplateFile(ctx, index.FindTemplate("demo"))
	assert.NoError(t, err)
	assert.Equal(t, fx.template, template)
}

func TestOfficialSourceRefusesAFileThatDoesNotMatchTheIndex(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.server.set("templates/demo.yaml", append(bytes.Clone(fx.template), []byte("# tampered\n")...))
	ctx := context.Background()

	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	entry := index.FindTemplate("demo")
	_, err = fx.source.TemplateFile(ctx, entry)

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVerificationFailed)
	assert.NoFileExists(t, filepath.Join(fx.cacheDir, entry.File.SHA256), "nothing unverified is cached")
}

func TestOfficialSourceRefusesAnIndexThatDoesNotMatchThePin(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.server.set(templaterepo.IndexFile, bytes.Replace(fx.indexData, []byte("Demo"), []byte("Evil"), 1))

	_, err := fx.source.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVerificationFailed)
}

func TestOfficialSourceRefetchesACorruptedCacheEntry(t *testing.T) {
	fx := newOfficialFixture(t)
	ctx := context.Background()
	index, err := fx.source.Index(ctx)
	assert.NoError(t, err)
	entry := index.FindTemplate("demo")
	_, err = fx.source.TemplateFile(ctx, entry)
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(fx.cacheDir, entry.File.SHA256), []byte("corrupt"), 0o600))
	template, err := fx.source.TemplateFile(ctx, entry)

	assert.NoError(t, err)
	assert.Equal(t, fx.template, template)
	assert.Equal(t, 2, fx.server.hitCount("templates/demo.yaml"))
}

func TestOfficialSourceRefusesAnOversizedFile(t *testing.T) {
	fx := newOfficialFixture(t)
	huge := []byte(strings.Repeat("x", templaterepo.MaxIndexSize+1))
	fx.release.Stable = pinOf(testCommit, huge)
	fx.server.set(templaterepo.IndexFile, huge)

	_, err := fx.source.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateFileTooLarge)
}

func TestOfficialSourceWithoutAPin(t *testing.T) {
	fx := newOfficialFixture(t)
	fx.release.Stable.Templates = nil

	_, err := fx.source.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplatesUnavailable)
}

func TestOfficialSourceReadsTheBetaPinInTheBetaEnv(t *testing.T) {
	fx := newOfficialFixture(t)
	useEnv(t, config.EnvBeta)
	fx.release.Beta = pinOf(testBetaCommit, fx.indexData)

	revision, err := fx.source.Revision(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, testBetaCommit, revision)
}

func TestOfficialSourceServesOnlyIconsTheIndexLists(t *testing.T) {
	fx := newOfficialFixture(t)
	index, err := fx.source.Index(context.Background())
	assert.NoError(t, err)

	_, err = fx.source.Icon(context.Background(), index.FindTemplate("demo").File.SHA256)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}
```

- [ ] **Step 5: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/`
Expected: FAIL - `undefined: officialSource`.

- [ ] **Step 6: Write the official source**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/source_official.go`:

```go
package apptemplateserviceimpl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	officialSourceID     = "official"
	rawGitHubBaseURL     = "https://raw.githubusercontent.com"
	fetchTimeout         = 15 * time.Second
	maxConcurrentFetches = 4
	cacheDirPerm         = 0o750
)

// officialSource serves the templates the signed release info pins.
//
// Nothing it returns is trusted because of where it came from. The release info
// is signed; the pin in it names the sha256 of index.json; the index names the
// sha256 of every template and icon; and every file is checked against its hash
// before anything reads it. The commit in the pin only chooses the URL - nothing
// here can check a git object id against what GitHub returns.
type officialSource struct {
	hpAppService hpappservice.Service
	baseURL      string
	client       *http.Client
	// cacheDir is where verified files are kept by hash, so a restart or an
	// unreachable GitHub still serves every revision already fetched.
	cacheDir   func() string
	fetchSlots chan struct{}

	mu        sync.Mutex
	indexSHA  string
	indexMemo *templatemodel.Index
}

func newOfficialSource(hpAppService hpappservice.Service) *officialSource {
	return &officialSource{
		hpAppService: hpAppService,
		baseURL:      rawGitHubBaseURL,
		client:       &http.Client{Timeout: fetchTimeout},
		cacheDir: func() string {
			return filepath.Join(config.Current().AppPath, "cache", "app-templates", "sha256")
		},
		fetchSlots: make(chan struct{}, maxConcurrentFetches),
	}
}

func (s *officialSource) ID() string {
	return officialSourceID
}

// pin reads the templates pin for the installation's channel from the release
// info, which is fetched, verified and cached by hpappservice.
func (s *officialSource) pin(ctx context.Context) (*base.TemplatesRef, error) {
	info, err := s.hpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	release := info.Stable
	if cfg := config.Current(); cfg != nil && cfg.IsBetaEnv() {
		release = info.Beta
	}
	if release == nil || release.Templates == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).
			WithExtraDetail("the release info pins no templates for this channel")
	}
	return release.Templates, nil
}

func (s *officialSource) Revision(ctx context.Context) (string, error) {
	pin, err := s.pin(ctx)
	if err != nil {
		return "", err
	}
	return pin.Commit, nil
}

func (s *officialSource) Index(ctx context.Context) (*templatemodel.Index, error) {
	pin, err := s.pin(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.indexMemo != nil && s.indexSHA == pin.IndexSHA256 {
		index := s.indexMemo
		s.mu.Unlock()
		return index, nil
	}
	s.mu.Unlock()

	data, err := s.fetchVerified(ctx, pin, templaterepo.IndexFile, pin.IndexSHA256, templaterepo.MaxIndexSize)
	if err != nil {
		return nil, err
	}
	index, err := templatemodel.DecodeIndex(data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	s.mu.Lock()
	s.indexMemo, s.indexSHA = index, pin.IndexSHA256
	s.mu.Unlock()
	return index, nil
}

func (s *officialSource) TemplateFile(ctx context.Context, entry *templatemodel.IndexEntry) ([]byte, error) {
	pin, err := s.pin(ctx)
	if err != nil {
		return nil, err
	}
	return s.fetchVerified(ctx, pin, entry.File.Path, entry.File.SHA256, templaterepo.MaxTemplateSize)
}

func (s *officialSource) Icon(ctx context.Context, sha256Hex string) ([]byte, error) {
	index, err := s.Index(ctx)
	if err != nil {
		return nil, err
	}
	entry := index.FindIcon(sha256Hex)
	if entry == nil {
		return nil, hperrors.NewNotFound("Icon")
	}
	pin, err := s.pin(ctx)
	if err != nil {
		return nil, err
	}
	return s.fetchVerified(ctx, pin, entry.Icon.Path, entry.Icon.SHA256, templaterepo.MaxIconSize)
}

// fetchVerified returns a file of the pinned revision whose sha256 must be want,
// from the cache when it has it and from GitHub otherwise.
func (s *officialSource) fetchVerified(
	ctx context.Context,
	pin *base.TemplatesRef,
	path, want string,
	maxSize int64,
) ([]byte, error) {
	if data, ok := s.readCache(want); ok {
		return data, nil
	}

	url := fmt.Sprintf("%s/%s/%s/%s", s.baseURL, pin.Repo, pin.Commit, path)
	data, err := s.download(ctx, url, path, maxSize)
	if err != nil {
		return nil, err
	}
	if hashHex(data) != want {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateVerificationFailed).
			WithExtraDetail("%s does not match its sha256", path)
	}
	s.writeCache(want, data)
	return data, nil
}

func (s *officialSource) download(ctx context.Context, url, path string, maxSize int64) ([]byte, error) {
	select {
	case s.fetchSlots <- struct{}{}:
	case <-ctx.Done():
		return nil, hperrors.Wrap(ctx.Err())
	}
	defer func() { <-s.fetchSlots }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).WithExtraDetail("%s: %s", path, err.Error())
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).WithExtraDetail("%s: %s", path, res.Status)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxSize+1))
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).WithExtraDetail("%s: %s", path, err.Error())
	}
	if int64(len(data)) > maxSize {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateFileTooLarge).
			WithExtraDetail("%s is larger than %d bytes", path, maxSize)
	}
	return data, nil
}

// readCache returns a cached file only if it still matches its hash. A file that
// does not is removed, and fetched again by the caller.
func (s *officialSource) readCache(sha256Hex string) ([]byte, bool) {
	path := filepath.Join(s.cacheDir(), sha256Hex)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	if hashHex(data) != sha256Hex {
		_ = os.Remove(path)
		return nil, false
	}
	return data, true
}

// writeCache stores a verified file by its hash. A write that fails costs a
// later refetch, not the request.
func (s *officialSource) writeCache(sha256Hex string, data []byte) {
	dir := s.cacheDir()
	if err := os.MkdirAll(dir, cacheDirPerm); err != nil {
		return
	}
	tmp, err := os.CreateTemp(dir, sha256Hex+".*.tmp")
	if err != nil {
		return
	}
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(tmp.Name())
		return
	}
	if err = os.Rename(tmp.Name(), filepath.Join(dir, sha256Hex)); err != nil {
		_ = os.Remove(tmp.Name())
	}
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
```

- [ ] **Step 7: Run the source tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/ -run OfficialSource`
Expected: PASS.

- [ ] **Step 8: Write the failing service tests**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service_test.go`:

```go
package apptemplateserviceimpl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type recordingLogger struct {
	logging.Logger
	warnings []string
}

func (l *recordingLogger) Warnf(template string, _ ...any) {
	l.warnings = append(l.warnings, template)
}

// unavailableSource is the official source when GitHub cannot be reached.
type unavailableSource struct{}

func (unavailableSource) ID() string { return officialSourceID }
func (unavailableSource) Revision(context.Context) (string, error) {
	return "", hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}
func (unavailableSource) Index(context.Context) (*templatemodel.Index, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}
func (unavailableSource) TemplateFile(context.Context, *templatemodel.IndexEntry) ([]byte, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}
func (unavailableSource) Icon(context.Context, string) ([]byte, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}

func newServiceTest(t *testing.T, env, dir string) (*service, *recordingLogger) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: env, AppTemplates: config.AppTemplates{Dir: dir}})
	t.Cleanup(func() { config.SetCurrent(previous) })

	logger := &recordingLogger{}
	return &service{official: unavailableSource{}, logger: logger}, logger
}

// copyTestRepo copies testdata/repo so a test can change a template.
func copyTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	assert.NoError(t, os.CopyFS(dir, os.DirFS(testRepoDir)))
	return dir
}

func TestServiceReadsTheLocalDirectoryInDevelopment(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	index, err := svc.Index(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, localSourceID, index.Source)
	assert.NotNil(t, index.Index.FindTemplate("demo"))
}

func TestServiceIgnoresTheLocalDirectoryOutsideDevelopment(t *testing.T) {
	svc, logger := newServiceTest(t, config.EnvProd, testRepoDir)

	_, err := svc.Index(context.Background())
	_, _ = svc.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplatesUnavailable, "the signed source is used instead")
	assert.Len(t, logger.warnings, 1, "the warning is logged once, not per request")
}

func TestServiceRender(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name:   "demo",
		Params: map[string]any{"dataVolume": "vol-1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "demo:2.1.0", resp.Result.Image)
	assert.Equal(t, localSourceID, resp.Source)
	assert.Equal(t, "templates/demo.yaml", resp.Entry.File.Path)
	assert.NotEmpty(t, resp.Entry.File.SHA256)
}

func TestServiceRenderRefusesAnIncompatibleTemplate(t *testing.T) {
	dir := copyTestRepo(t)
	path := filepath.Join(dir, "templates", "demo.yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(content), "v000001", "v999999", 1)), 0o600))
	svc, _ := newServiceTest(t, config.EnvDev, dir)

	_, err = svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name: "demo", Params: map[string]any{"dataVolume": "vol-1"},
	})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateIncompatible)
}

func TestServiceTemplateNotFound(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	_, err := svc.Template(context.Background(), "nope")

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateNotFound)
}

func TestServiceIcon(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	index, err := svc.Index(context.Background())
	assert.NoError(t, err)

	icon, err := svc.Icon(context.Background(), index.Index.FindTemplate("demo").Icon.SHA256)

	assert.NoError(t, err)
	assert.Equal(t, "image/svg+xml", icon.ContentType)
	assert.Contains(t, string(icon.Content), "<svg")
}
```

- [ ] **Step 9: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/ -run Service`
Expected: FAIL - `undefined: service`, `undefined: localSourceID`.

- [ ] **Step 10: Write the local directory source and the service**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/source_localdir.go`:

```go
package apptemplateserviceimpl

import (
	"context"
	"os"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

const (
	localSourceID = "local"
	localRevision = "local"
)

// localDirSource reads a checkout of the templates repository directly: no
// signature, no hashes, re-read on every call so an edit shows up at once. It
// exists for authoring templates, and the service only uses it in development.
type localDirSource struct {
	dir string
}

func newLocalDirSource(dir string) *localDirSource {
	return &localDirSource{dir: dir}
}

func (s *localDirSource) ID() string {
	return localSourceID
}

func (s *localDirSource) Revision(context.Context) (string, error) {
	return localRevision, nil
}

// load reads the checkout. A template that does not decode is left out, as it
// would be left out of an index; tools/apptemplate lint says why.
func (s *localDirSource) load() (*templaterepo.Repo, *templatemodel.Index, error) {
	repo, _, err := templaterepo.Load(os.DirFS(s.dir))
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	index, err := templaterepo.BuildIndex(repo)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return repo, index, nil
}

func (s *localDirSource) Index(context.Context) (*templatemodel.Index, error) {
	_, index, err := s.load()
	return index, err
}

func (s *localDirSource) TemplateFile(_ context.Context, entry *templatemodel.IndexEntry) ([]byte, error) {
	repo, _, err := s.load()
	if err != nil {
		return nil, err
	}
	file := repo.FindTemplate(entry.Name)
	if file == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateNotFound).WithParam("Name", entry.Name)
	}
	return file.Content, nil
}

func (s *localDirSource) Icon(_ context.Context, sha256Hex string) ([]byte, error) {
	repo, index, err := s.load()
	if err != nil {
		return nil, err
	}
	entry := index.FindIcon(sha256Hex)
	if entry == nil {
		return nil, hperrors.NewNotFound("Icon")
	}
	return repo.Icons[entry.Icon.Path].Content, nil
}
```

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go`:

```go
package apptemplateserviceimpl

import (
	"context"
	"path"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	contentTypeSVG = "image/svg+xml"
	contentTypePNG = "image/png"
)

func New(
	hpAppService hpappservice.Service,
	logger logging.Logger,
) apptemplateservice.Service {
	return &service{
		official: newOfficialSource(hpAppService),
		logger:   logger,
	}
}

type service struct {
	official apptemplateservice.Source
	logger   logging.Logger
	warnOnce sync.Once
}

// source picks where templates come from. HP_TEMPLATES_DIR is read with no
// signature and no hashes, so it is honored only in development - anywhere else
// it would let whoever sets an environment variable choose what images run.
func (s *service) source() apptemplateservice.Source {
	cfg := config.Current()
	if cfg == nil || cfg.AppTemplates.Dir == "" {
		return s.official
	}
	if cfg.IsDevEnv() {
		return newLocalDirSource(cfg.AppTemplates.Dir)
	}
	s.warnOnce.Do(func() {
		s.logger.Warnf("app templates: HP_TEMPLATES_DIR is set but ignored outside the development environment")
	})
	return s.official
}

func (s *service) Index(ctx context.Context) (*apptemplateservice.IndexResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	revision, err := src.Revision(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.IndexResp{Source: src.ID(), Revision: revision, Index: index}, nil
}

func (s *service) Template(ctx context.Context, name string) (*apptemplateservice.TemplateResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry := index.FindTemplate(name)
	if entry == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateNotFound).WithParam("Name", name)
	}
	data, err := src.TemplateFile(ctx, entry)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	tmpl, err := templatemodel.DecodeTemplate(data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	revision, err := src.Revision(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.TemplateResp{Source: src.ID(), Revision: revision, Entry: entry, Template: tmpl}, nil
}

func (s *service) Icon(ctx context.Context, sha256Hex string) (*apptemplateservice.IconResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry := index.FindIcon(sha256Hex)
	if entry == nil {
		return nil, hperrors.NewNotFound("Icon")
	}
	content, err := src.Icon(ctx, sha256Hex)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	contentType := contentTypePNG
	if path.Ext(entry.Icon.Path) == ".svg" {
		contentType = contentTypeSVG
	}
	return &apptemplateservice.IconResp{Content: content, ContentType: contentType}, nil
}

func (s *service) Render(
	ctx context.Context,
	req *apptemplateservice.RenderReq,
) (*apptemplateservice.RenderResp, error) {
	loaded, err := s.Template(ctx, req.Name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !templatemodel.IsCompatible(loaded.Template.Metadata.Requires, base.CurrentVersion) {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).WithParam("Name", req.Name)
	}
	result, err := templaterender.Render(&templaterender.Request{
		Template: loaded.Template,
		Version:  req.Version,
		Variant:  req.Variant,
		Params:   req.Params,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.RenderResp{TemplateResp: *loaded, Result: result}, nil
}
```

In `hivepaas_app/registry/provides.go`, add `apptemplateserviceimpl.New,` to the service constructors after `appprovisionserviceimpl.New,`, with its import.

- [ ] **Step 11: Run the package tests**

Run: `go build ./... && go test ./hivepaas_app/service/apptemplateservice/...`
Expected: PASS.

- [ ] **Step 12: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice hivepaas_app/config/app_templates.go hivepaas_app/config/config.go \
  hivepaas_app/hperrors/errors_app_template.go hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml \
  hivepaas_app/registry/provides.go
git commit -m "feat(apptemplate): fetch, verify and cache templates from the signed pin"
```

---
## Task 12: The catalogue API - `apptemplateuc` reads

**Files:**
- Create: `hivepaas_app/usecase/apptemplateuc/uc.go`
- Create: `hivepaas_app/usecase/apptemplateuc/catalog.go`
- Create: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/list.go`
- Create: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/get.go`
- Create: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/icon_get.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/handler.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/list.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/get.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/icon_get.go`
- Create: `hivepaas_app/interface/api/server/router_app_templates.go`
- Modify: `hivepaas_app/interface/api/server/router.go`, `router_projects.go`
- Modify: `hivepaas_app/registry/provides.go`
- Modify: `docs/openapi/swagger.json` (generated)
- Test: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/transform_test.go`

**Interfaces:**
- Consumes: `apptemplateservice.Service`, `IndexResp`, `TemplateResp`, `IconResp` (Task 11); `templatemodel` (Task 1).
- Produces:
  - `apptemplateuc.New(appTemplateService apptemplateservice.Service) *UC` - Task 13 adds the dependencies creating an app needs
  - `(*UC).ListAppTemplates(ctx, auth, *ListAppTemplatesReq) (*ListAppTemplatesResp, error)`
  - `(*UC).GetAppTemplate(ctx, auth, *GetAppTemplateReq) (*GetAppTemplateResp, error)`
  - `(*UC).GetAppTemplateIcon(ctx, *GetAppTemplateIconReq) (*GetAppTemplateIconResp, error)`
  - `apptemplatedto.TransformAppTemplateCatalog(index *apptemplateservice.IndexResp, currentVersionCode string) *AppTemplateCatalogResp`
  - `apptemplatedto.TransformAppTemplate(tmpl *apptemplateservice.TemplateResp, currentVersionCode string) *AppTemplateResp`
  - `apptemplatehandler.Handler` with `ListAppTemplates`, `GetAppTemplate`, `GetAppTemplateIcon`
  - routes `GET /projects/:projectID/app-templates`, `GET /projects/:projectID/app-templates/:templateName`, `GET /app-templates/icons/:sha256` (public)

- [ ] **Step 1: Write the failing transform tests**

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/transform_test.go`:

```go
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
				{Name: "memoryLimit", Title: "Memory", Type: templatemodel.ParamTypeSize, Default: "512MB", Min: "128MB"},
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
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/...`
Expected: FAIL - `undefined: TransformAppTemplateCatalog`.

- [ ] **Step 3: Write the DTOs**

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/list.go`:

```go
package apptemplatedto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type ListAppTemplatesReq struct {
	ProjectID string `json:"-"`
}

func NewListAppTemplatesReq() *ListAppTemplatesReq {
	return &ListAppTemplatesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ListAppTemplatesReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ProjectID, true, "projectId")...))
}

type ListAppTemplatesResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppTemplateCatalogResp `json:"data"`
}

type AppTemplateCatalogResp struct {
	Source     string                     `json:"source"`
	Revision   string                     `json:"revision"`
	Categories []*AppTemplateCategoryResp `json:"categories"`
	Tags       []*AppTemplateTagResp      `json:"tags"`
	Templates  []*AppTemplateSummaryResp  `json:"templates"`
}

type AppTemplateCategoryResp struct {
	ID       string                     `json:"id"`
	Title    string                     `json:"title"`
	Children []*AppTemplateCategoryResp `json:"children,omitempty"`
}

type AppTemplateTagResp struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type AppTemplateSummaryResp struct {
	Name       string                           `json:"name"`
	Title      string                           `json:"title"`
	Tagline    string                           `json:"tagline"`
	Categories []string                         `json:"categories"`
	Tags       []string                         `json:"tags"`
	Aliases    []string                         `json:"aliases"`
	IconURL    string                           `json:"iconUrl"`
	Variants   []*AppTemplateVariantSummaryResp `json:"variants"`
	Versions   []*AppTemplateVersionResp        `json:"versions"`
	// Compatible is false for a template needing a newer HivePaaS: the store
	// lists it, locked, rather than hiding it.
	Compatible          bool   `json:"compatible"`
	RequiresVersionCode string `json:"requiresVersionCode"`
}

type AppTemplateVariantSummaryResp struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type AppTemplateVersionResp struct {
	Name       string   `json:"name"`
	Release    string   `json:"release"`
	Default    bool     `json:"default"`
	Deprecated bool     `json:"deprecated"`
	Variants   []string `json:"variants"`
}

func TransformAppTemplateCatalog(
	index *apptemplateservice.IndexResp,
	currentVersionCode string,
) *AppTemplateCatalogResp {
	resp := &AppTemplateCatalogResp{
		Source:     index.Source,
		Revision:   index.Revision,
		Categories: transformCategories(index.Index.Categories),
		Tags:       make([]*AppTemplateTagResp, 0, len(index.Index.Tags)),
		Templates:  make([]*AppTemplateSummaryResp, 0, len(index.Index.Templates)),
	}
	for _, tag := range index.Index.Tags {
		resp.Tags = append(resp.Tags, &AppTemplateTagResp{ID: tag.ID, Title: tag.Title})
	}
	for _, entry := range index.Index.Templates {
		resp.Templates = append(resp.Templates, transformSummary(entry, currentVersionCode))
	}
	return resp
}

func transformCategories(categories []*templatemodel.Category) []*AppTemplateCategoryResp {
	out := make([]*AppTemplateCategoryResp, 0, len(categories))
	for _, category := range categories {
		out = append(out, &AppTemplateCategoryResp{
			ID:       category.ID,
			Title:    category.Title,
			Children: transformCategories(category.Children),
		})
	}
	return out
}

func transformSummary(entry *templatemodel.IndexEntry, currentVersionCode string) *AppTemplateSummaryResp {
	summary := &AppTemplateSummaryResp{
		Name:                entry.Name,
		Title:               entry.Title,
		Tagline:             entry.Tagline,
		Categories:          entry.Categories,
		Tags:                entry.Tags,
		Aliases:             entry.Aliases,
		IconURL:             AppTemplateIconURL(entry.Icon.SHA256),
		Variants:            make([]*AppTemplateVariantSummaryResp, 0, len(entry.Variants)),
		Versions:            make([]*AppTemplateVersionResp, 0, len(entry.Versions)),
		Compatible:          templatemodel.IsCompatible(entry.Requires, currentVersionCode),
		RequiresVersionCode: entry.Requires.VersionCode,
	}
	for _, variant := range entry.Variants {
		summary.Variants = append(summary.Variants,
			&AppTemplateVariantSummaryResp{Name: variant.Name, Default: variant.Default})
	}
	for _, version := range entry.Versions {
		summary.Versions = append(summary.Versions, &AppTemplateVersionResp{
			Name:       version.Name,
			Release:    version.Release,
			Default:    version.Default,
			Deprecated: version.Deprecated,
			Variants:   version.Variants,
		})
	}
	return summary
}

// AppTemplateIconURL is where the dashboard loads an icon from. The route is
// public, because an <img> cannot send the Authorization header, and addressed by
// hash, so the response can be cached forever.
func AppTemplateIconURL(sha256Hex string) string {
	return fmt.Sprintf("%v/app-templates/icons/%v", config.Current().HTTPServer.BasePath, sha256Hex)
}
```

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/get.go`:

```go
package apptemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const templateNameMaxLen = 63

type GetAppTemplateReq struct {
	ProjectID string `json:"-"`
	Name      string `json:"-"`
}

func NewGetAppTemplateReq() *GetAppTemplateReq {
	return &GetAppTemplateReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, templateNameMaxLen, "name")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppTemplateResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *AppTemplateResp `json:"data"`
}

type AppTemplateResp struct {
	Source              string                    `json:"source"`
	Revision            string                    `json:"revision"`
	Name                string                    `json:"name"`
	Title               string                    `json:"title"`
	Tagline             string                    `json:"tagline"`
	Description         string                    `json:"description"`
	Categories          []string                  `json:"categories"`
	Tags                []string                  `json:"tags"`
	IconURL             string                    `json:"iconUrl"`
	Links               *AppTemplateLinksResp     `json:"links"`
	License             string                    `json:"license"`
	Compatible          bool                      `json:"compatible"`
	RequiresVersionCode string                    `json:"requiresVersionCode"`
	Variants            []*AppTemplateVariantResp `json:"variants"`
	Versions            []*AppTemplateVersionResp `json:"versions"`
	Parameters          []*AppTemplateParamResp   `json:"parameters"`
}

type AppTemplateLinksResp struct {
	Website       string `json:"website,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	Source        string `json:"source,omitempty"`
}

type AppTemplateVariantResp struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

type AppTemplateParamResp struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	// Default is never set for a secret.
	Default   any  `json:"default,omitempty"`
	Optional  bool `json:"optional"`
	Pattern   string `json:"pattern,omitempty"`
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Min       any    `json:"min,omitempty"`
	Max       any    `json:"max,omitempty"`
	// Generated says a secret left empty is generated.
	Generated bool                          `json:"generated"`
	Options   []*AppTemplateParamOptionResp `json:"options,omitempty"`
}

type AppTemplateParamOptionResp struct {
	Value string `json:"value"`
	Title string `json:"title"`
}

func TransformAppTemplate(tmpl *apptemplateservice.TemplateResp, currentVersionCode string) *AppTemplateResp {
	metadata := tmpl.Template.Metadata
	summary := transformSummary(tmpl.Entry, currentVersionCode)
	resp := &AppTemplateResp{
		Source:              tmpl.Source,
		Revision:            tmpl.Revision,
		Name:                metadata.Name,
		Title:               metadata.Title,
		Tagline:             metadata.Tagline,
		Description:         metadata.Description,
		Categories:          metadata.Categories,
		Tags:                metadata.Tags,
		IconURL:             summary.IconURL,
		License:             metadata.License,
		Compatible:          summary.Compatible,
		RequiresVersionCode: summary.RequiresVersionCode,
		Versions:            summary.Versions,
		Variants:            make([]*AppTemplateVariantResp, 0, len(tmpl.Template.Variants)),
		Parameters:          make([]*AppTemplateParamResp, 0, len(tmpl.Template.Parameters)),
	}
	if links := metadata.Links; links != nil {
		resp.Links = &AppTemplateLinksResp{
			Website: links.Website, Documentation: links.Documentation, Source: links.Source,
		}
	}
	for _, variant := range tmpl.Template.Variants {
		resp.Variants = append(resp.Variants, &AppTemplateVariantResp{
			Name: variant.Name, Title: variant.Title, Description: variant.Description, Default: variant.Default,
		})
	}
	for _, param := range tmpl.Template.Parameters {
		resp.Parameters = append(resp.Parameters, transformParam(param))
	}
	return resp
}

func transformParam(param *templatemodel.Parameter) *AppTemplateParamResp {
	resp := &AppTemplateParamResp{
		Name:        param.Name,
		Title:       param.Title,
		Description: param.Description,
		Type:        string(param.Type),
		Default:     param.Default,
		Optional:    param.Optional,
		Pattern:     param.Pattern,
		MinLength:   param.MinLength,
		MaxLength:   param.MaxLength,
		Min:         param.Min,
		Max:         param.Max,
		Generated:   param.Generate != nil,
	}
	if param.Type == templatemodel.ParamTypeSecret {
		// Validation refuses a secret default already; this keeps one from
		// reaching a response if a template ever slips past it.
		resp.Default = nil
	}
	for _, option := range param.Options {
		resp.Options = append(resp.Options, &AppTemplateParamOptionResp{Value: option.Value, Title: option.Title})
	}
	return resp
}
```

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/icon_get.go`:

```go
package apptemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const sha256HexLen = 64

type GetAppTemplateIconReq struct {
	SHA256 string `json:"-"`
}

func NewGetAppTemplateIconReq() *GetAppTemplateIconReq {
	return &GetAppTemplateIconReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateIconReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateStr(&req.SHA256, true, sha256HexLen, sha256HexLen, "sha256")...))
}

// GetAppTemplateIconResp is written as the image itself, not as JSON.
type GetAppTemplateIconResp struct {
	Content     []byte
	ContentType string
}
```

- [ ] **Step 4: Run the transform tests**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/...`
Expected: PASS.

- [ ] **Step 5: Write the usecase**

`hivepaas_app/usecase/apptemplateuc/uc.go`:

```go
package apptemplateuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
)

type UC struct {
	appTemplateService apptemplateservice.Service
}

func New(
	appTemplateService apptemplateservice.Service,
) *UC {
	return &UC{
		appTemplateService: appTemplateService,
	}
}
```

`hivepaas_app/usecase/apptemplateuc/catalog.go`:

```go
package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) ListAppTemplates(
	ctx context.Context,
	_ *basedto.Auth,
	_ *apptemplatedto.ListAppTemplatesReq,
) (*apptemplatedto.ListAppTemplatesResp, error) {
	index, err := uc.appTemplateService.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.ListAppTemplatesResp{
		Data: apptemplatedto.TransformAppTemplateCatalog(index, base.CurrentVersion),
	}, nil
}

func (uc *UC) GetAppTemplate(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateReq,
) (*apptemplatedto.GetAppTemplateResp, error) {
	tmpl, err := uc.appTemplateService.Template(ctx, req.Name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateResp{
		Data: apptemplatedto.TransformAppTemplate(tmpl, base.CurrentVersion),
	}, nil
}

func (uc *UC) GetAppTemplateIcon(
	ctx context.Context,
	req *apptemplatedto.GetAppTemplateIconReq,
) (*apptemplatedto.GetAppTemplateIconResp, error) {
	icon, err := uc.appTemplateService.Icon(ctx, req.SHA256)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateIconResp{Content: icon.Content, ContentType: icon.ContentType}, nil
}
```

- [ ] **Step 6: Write the handlers and routes**

`hivepaas_app/interface/api/handler/apptemplatehandler/handler.go`:

```go
package apptemplatehandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc"
)

type Handler struct {
	*appbasehandler.Handler
	appTemplateUC *apptemplateuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appTemplateUC *apptemplateuc.UC,
) *Handler {
	return &Handler{
		Handler:       baseHandler,
		appTemplateUC: appTemplateUC,
	}
}
```

`hivepaas_app/interface/api/handler/apptemplatehandler/list.go`:

```go
package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// ListAppTemplates Lists the app template catalogue
// @Summary Lists the app template catalogue
// @Description Lists categories, tags and a summary of every app template a project can create apps from.
// @Tags    app_templates
// @Produce json
// @Id      listAppTemplates
// @Param   projectID path string true "project ID"
// @Success 200 {object} apptemplatedto.ListAppTemplatesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/app-templates [get]
func (h *Handler) ListAppTemplates(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewListAppTemplatesReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.ListAppTemplates(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
```

`hivepaas_app/interface/api/handler/apptemplatehandler/get.go`:

```go
package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplate Gets an app template
// @Summary Gets an app template
// @Description Gets one app template: its description, versions, variants and parameters.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplate
// @Param   projectID path string true "project ID"
// @Param   templateName path string true "template name"
// @Success 200 {object} apptemplatedto.GetAppTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/app-templates/{templateName} [get]
func (h *Handler) GetAppTemplate(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	name, err := h.ParseStringParam(ctx, "templateName")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateReq()
	req.ProjectID = projectID
	req.Name = name
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplate(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
```

`hivepaas_app/interface/api/handler/apptemplatehandler/icon_get.go`:

```go
package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateIcon Gets an app template icon
// @Summary Gets an app template icon
// @Description Public: an <img> cannot send the Authorization header. Only icons the current index lists are served.
// @Tags    app_templates
// @Produce image/svg+xml,image/png
// @Id      getAppTemplateIcon
// @Param   sha256 path string true "icon sha256"
// @Success 200
// @Failure 404 {object} hperrors.ErrorInfo
// @Router  /app-templates/icons/{sha256} [get]
func (h *Handler) GetAppTemplateIcon(ctx *gin.Context) {
	sha256Hex, err := h.ParseStringParam(ctx, "sha256")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateIconReq()
	req.SHA256 = sha256Hex
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplateIcon(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// An SVG can carry script. sandbox keeps it inert even when opened directly,
	// and nosniff keeps the declared type the only one a browser uses.
	ctx.Header("Content-Security-Policy", "sandbox")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Data(http.StatusOK, resp.ContentType, resp.Content)
}
```

`hivepaas_app/interface/api/server/router_app_templates.go`:

```go
package server

import (
	"github.com/gin-gonic/gin"
)

// registerAppTemplateRoutes registers the public icon route. The catalogue itself
// is read under a project - see registerProjectRoutes - and creating an app from
// a template under its env - see registerAppRoutes.
func (s *HTTPServer) registerAppTemplateRoutes(apiGroup *gin.RouterGroup) {
	apiGroup.GET("/app-templates/icons/:sha256", s.handlerRegistry.appTemplateHandler.GetAppTemplateIcon)
}
```

In `hivepaas_app/interface/api/server/router.go`:
- add `appTemplateHandler *apptemplatehandler.Handler` to `HandlerRegistry` after `appSettingsHandler`, the same-named parameter to `NewHandlerRegistry` after `appSettingsHandler`, its assignment, and the import;
- call `s.registerAppTemplateRoutes(apiGroup)` right after `s.registerImageRoutes(apiGroup)`.

In `hivepaas_app/interface/api/server/router_projects.go`, after the `// Configuration spec export` line and its route:

```go
	// App templates catalogue
	projectGroup.GET("/:projectID/app-templates", s.handlerRegistry.appTemplateHandler.ListAppTemplates)
	projectGroup.GET("/:projectID/app-templates/:templateName", s.handlerRegistry.appTemplateHandler.GetAppTemplate)
```

In `hivepaas_app/registry/provides.go`, add `apptemplatehandler.New,` after `appsettingshandler.New,` and `apptemplateuc.New,` beside the other usecase constructors, with their imports.

- [ ] **Step 7: Build, generate swagger, and try it**

Run: `go build ./... && go test ./hivepaas_app/usecase/apptemplateuc/... && make gen-swag`
Expected: PASS, and `git diff --stat docs/openapi/swagger.json` shows the three new operations.

Then try it on a second backend reading the test repository, on its own port so an instance already on port 10000 is left alone (`HP_ENV=development` is what `config/config.local.toml` sets; the dev sign-in is the one in `docs/DEVELOPMENT.md` §6):

```bash
HP_HTTP_SERVER_PORT=10077 HP_RUN_MODE=app \
  HP_TEMPLATES_DIR="$PWD/hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo" \
  make local-app-run &

USER_ID=$(PGPASSWORD=abc123 psql -h localhost -p 35432 -U hivepaas -d hivepaas \
  -tAc "SELECT id FROM users WHERE deleted_at IS NULL ORDER BY created_at LIMIT 1")
TOKEN=$(curl -sS -u hivepaas:abc123 -X POST \
  "http://localhost:10077/_/internal/dev-helper/dev-mode-login?userId=$USER_ID" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["accessToken"])')

curl -sS "http://localhost:10077/_/projects/<projectID>/app-templates" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import json,sys; d=json.load(sys.stdin)["data"]; print(d["source"], d["templates"][0]["iconUrl"])'
curl -sSI "http://localhost:10077<iconUrl printed above>"
```

Expected: `local /_/app-templates/icons/<sha256>`; the icon answers `200` without an `Authorization` header, with `Content-Type: image/svg+xml` and `Content-Security-Policy: sandbox`. Stop the second backend afterwards.

- [ ] **Step 8: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/usecase/apptemplateuc hivepaas_app/interface/api/handler/apptemplatehandler \
  hivepaas_app/interface/api/server hivepaas_app/registry/provides.go docs/openapi/swagger.json
git commit -m "feat(apptemplate): catalogue, template detail and icon endpoints"
```

---
## Task 13: Create an app from a template, and read an app's template

**Files:**
- Modify: `hivepaas_app/usecase/apptemplateuc/uc.go`
- Create: `hivepaas_app/usecase/apptemplateuc/create.go`
- Create: `hivepaas_app/usecase/apptemplateuc/binding_get.go`
- Create: `hivepaas_app/usecase/apptemplateuc/audit.go`
- Create: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/create.go`
- Create: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/binding_get.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/create.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/binding_get.go`
- Modify: `hivepaas_app/interface/api/server/router_apps.go`
- Modify: `docs/openapi/swagger.json` (generated)
- Test: `hivepaas_app/usecase/apptemplateuc/create_test.go`

**Interfaces:**
- Consumes: `apptemplateservice.Service.Render`, `RenderReq`, `RenderResp` (Task 11); `templaterender.Result`, `Value` (Tasks 4-5); `templatemodel.ParamTypeSecret` (Task 1); `specservice.Service.BuildApp`, `BuildAppReq` (Task 9); `appprovisionservice.Service.ProvisionApp`, `ProvisionAppReq` (Task 10); `entity.AppTemplateSettings`, `AppTemplateParam`, `AppTemplateBase`, `CurrentAppTemplateSettingsVersion`, `base.SettingTypeAppTemplate` (Task 7); existing `appdeploymentservice.CreateDeploymentAndTask`, `envvarservice`, `approutingservice`, `appservice.PersistAppData`, `appservice.LoadApp`, `auditservice.RecordAllowed`, `queue.TaskQueue.ScheduleTask`.
- Produces:
  - `apptemplateuc.New(db *database.DB, taskQueue queue.TaskQueue, settingRepo repository.SettingRepo, appDeploymentService appdeploymentservice.Service, appProvisionService appprovisionservice.Service, appRoutingService approutingservice.Service, appService appservice.Service, appTemplateService apptemplateservice.Service, auditService auditservice.Service, clusterService clusterservice.Service, envVarService envvarservice.Service, specService specservice.Service) *UC`
  - `(*UC).CreateAppFromTemplate(ctx, auth, *CreateAppFromTemplateReq) (*CreateAppFromTemplateResp, error)`
  - `(*UC).GetAppTemplateBinding(ctx, auth, *GetAppTemplateBindingReq) (*GetAppTemplateBindingResp, error)`
  - `apptemplatedto.CreateAppFromTemplateReq{ProjectID, ProjectEnvID string (json:"-"); Name, Template, Version, Variant string; Params map[string]any}`
  - `apptemplatedto.CreateAppFromTemplateResp{Meta; Data *CreateAppFromTemplateDataResp{App, Deployment *basedto.ObjectIDResp}}` - links are objects, per ARCHITECTURE §2
  - `apptemplatedto.AppTemplateBindingResp{Source, Template, Title, Version, Release, Variant, Revision string; AppliedAt time.Time}`
  - routes `POST /projects/:projectID/:projectEnv/apps/from-template`, `GET /projects/:projectID/:projectEnv/apps/:appID/template`

Rendering happens before the transaction: it may reach GitHub, and nothing should be locked while it does. Everything inside the transaction lives in `provisionFromTemplate`, which takes `database.IDB` so it can be tested with a nil database - the same seam `hpappuc` uses.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/usecase/apptemplateuc/create_test.go`:

```go
package apptemplateuc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

const testTemplateYAML = `
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
  - {name: username, title: Username, type: string, default: app}
  - {name: password, title: Password, type: secret, generate: {length: 24}}
  - {name: dataVolume, title: Data volume, type: volume}
variants:
  - {name: alpine, title: Alpine, default: true}
versions:
  - {name: "18", release: "18.6", default: true, images: {alpine: "postgres:18.6-alpine3.24"}}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
  settings:
    kind:
      category: database
      engine: postgres
      database: {username: "${{ params.username }}", password: "${{ params.password }}"}
    routing: {port: 5432}
`

var errTestRouting = errors.New("traefik unavailable")

type fakeTemplateService struct {
	apptemplateservice.Service
	resp *apptemplateservice.RenderResp
	err  error
}

func (f *fakeTemplateService) Render(
	context.Context, *apptemplateservice.RenderReq,
) (*apptemplateservice.RenderResp, error) {
	return f.resp, f.err
}

// fakeProvisionService runs Configure the way ProvisionApp does and hands back
// the app with what it returned.
type fakeProvisionService struct {
	appprovisionservice.Service
	called bool
}

func (f *fakeProvisionService) ProvisionApp(
	ctx context.Context, db database.IDB, req *appprovisionservice.ProvisionAppReq,
) (*appprovisionservice.ProvisionAppResp, error) {
	f.called = true
	app := &entity.App{
		ID: "app-1", Name: req.Name, Key: "main-db", ServiceID: "svc-1",
		ProjectID: req.ProjectID, ProjectEnvID: req.ProjectEnvID,
		Project:    &entity.Project{ID: req.ProjectID, Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{ID: req.ProjectEnvID, Key: "prod", Name: "prod"},
	}
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}}
	settings, err := req.Configure(ctx, db, app, spec)
	if err != nil {
		return nil, err
	}
	app.Settings = settings
	return &appprovisionservice.ProvisionAppResp{App: app}, nil
}

type fakeSpecService struct {
	specservice.Service
}

func (f *fakeSpecService) BuildApp(
	_ context.Context, _ database.IDB, req *specservice.BuildAppReq,
) (*specservice.BuildAppResp, error) {
	deployment := &entity.Setting{ID: "set-deploy", Type: base.SettingTypeAppDeployment, ObjectID: req.App.ID}
	deployment.MustSetData(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodImage,
		ImageSource:  &entity.DeploymentImageSource{Image: "postgres:18.6-alpine3.24"},
	})
	routing := &entity.Setting{ID: "set-routing", Type: base.SettingTypeAppRouting, ObjectID: req.App.ID}
	routing.MustSetData(&entity.AppRoutingSettings{Port: 5432})
	return &specservice.BuildAppResp{Settings: []*entity.Setting{deployment, routing}}, nil
}

type fakeEnvVarService struct {
	envvarservice.Service
	applied bool
}

func (f *fakeEnvVarService) BuildEnvVarsForAllAppsInScope(
	context.Context, database.IDB, *entity.ObjectScope, bool, []string, bool, bool,
) ([]*envvarservice.AppEnvVarData, error) {
	return nil, nil
}

func (f *fakeEnvVarService) ApplyEnvVarsForApps(
	context.Context, database.IDB, []*envvarservice.AppEnvVarData, bool, bool,
) map[int]error {
	f.applied = true
	return nil
}

type fakeRoutingService struct {
	approutingservice.Service
	req *approutingservice.ApplyAppRoutingReq
	err error
}

func (f *fakeRoutingService) ApplyRoutingSettings(
	_ context.Context, _ database.IDB, req *approutingservice.ApplyAppRoutingReq,
) (*approutingservice.ApplyAppRoutingResp, error) {
	f.req = req
	return &approutingservice.ApplyAppRoutingResp{}, f.err
}

type fakeDeploymentService struct {
	appdeploymentservice.Service
}

func (f *fakeDeploymentService) CreateDeploymentAndTask(
	app *entity.App, settings *entity.AppDeploymentSettings,
) (*entity.Deployment, *entity.Task, error) {
	return &entity.Deployment{ID: "dep-1", AppID: app.ID, Settings: settings}, &entity.Task{ID: "task-1"}, nil
}

type fakeAppService struct {
	appservice.Service
	persisted *appservice.PersistingAppData
}

func (f *fakeAppService) PersistAppData(_ context.Context, _ database.IDB, data *appservice.PersistingAppData) error {
	f.persisted = data
	return nil
}

type fakeAuditService struct {
	auditservice.Service
	entries []*auditservice.Entry
}

func (f *fakeAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type createFakes struct {
	templates *fakeTemplateService
	provision *fakeProvisionService
	envVars   *fakeEnvVarService
	routing   *fakeRoutingService
	apps      *fakeAppService
	audit     *fakeAuditService
}

func newCreateTest(t *testing.T) (*UC, *createFakes) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)

	tmpl, err := templatemodel.DecodeTemplate([]byte(testTemplateYAML))
	assert.NoError(t, err)
	result, err := templaterender.Render(&templaterender.Request{
		Template: tmpl,
		Params:   map[string]any{"dataVolume": "vol-1"},
	})
	assert.NoError(t, err)

	fakes := &createFakes{
		templates: &fakeTemplateService{resp: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{
				Source:   "official",
				Revision: "0123456789abcdef0123456789abcdef01234567",
				Entry:    &templatemodel.IndexEntry{Name: "pg", File: templatemodel.FileRef{SHA256: "file-sha"}},
				Template: tmpl,
			},
			Result: result,
		}},
		provision: &fakeProvisionService{},
		envVars:   &fakeEnvVarService{},
		routing:   &fakeRoutingService{},
		apps:      &fakeAppService{},
		audit:     &fakeAuditService{},
	}
	uc := &UC{
		appDeploymentService: &fakeDeploymentService{},
		appProvisionService:  fakes.provision,
		appRoutingService:    fakes.routing,
		appService:           fakes.apps,
		appTemplateService:   fakes.templates,
		auditService:         fakes.audit,
		envVarService:        fakes.envVars,
		specService:          &fakeSpecService{},
	}
	return uc, fakes
}

func testCreateReq() *apptemplatedto.CreateAppFromTemplateReq {
	return &apptemplatedto.CreateAppFromTemplateReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "main-db", Template: "pg",
		Params: map[string]any{"dataVolume": "vol-1"},
	}
}

func testAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "user-1"}}}
}

func provision(t *testing.T, uc *UC, fakes *createFakes) *createdFromTemplate {
	t.Helper()
	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		fakes.templates.resp)
	assert.NoError(t, err)
	return created
}

func TestProvisionFromTemplateRecordsTheTemplate(t *testing.T) {
	uc, fakes := newCreateTest(t)

	created := provision(t, uc, fakes)

	var binding *entity.Setting
	for _, setting := range created.app.Settings {
		if setting.Type == base.SettingTypeAppTemplate {
			binding = setting
		}
	}
	if binding == nil {
		t.Fatal("the app records the template it was provisioned from")
	}
	assert.Equal(t, base.ObjectScopeApp, binding.Scope)
	assert.Equal(t, "app-1", binding.ObjectID)

	stored := &entity.Setting{Type: base.SettingTypeAppTemplate, Data: binding.Data}
	data, err := stored.AsAppTemplateSettings()
	assert.NoError(t, err)
	assert.Equal(t, "official", data.Source)
	assert.Equal(t, "pg", data.Template)
	assert.Equal(t, "PG", data.Title)
	assert.Equal(t, "18", data.Version)
	assert.Equal(t, "alpine", data.Variant)
	assert.Equal(t, "18.6", data.Base.Release)
	assert.Equal(t, "file-sha", data.Base.TemplateSHA256)
	assert.Equal(t, fakes.templates.resp.Result.BaseSHA256, data.Base.RenderedSHA256)
	assert.Equal(t, "app", data.Params["username"].Value)
	assert.Equal(t, "vol-1", data.Params["dataVolume"].Value)

	generated := fakes.templates.resp.Result.Params["password"].Text()
	assert.NotContains(t, binding.Data, generated, "a secret parameter is stored encrypted")
	assert.NotContains(t, data.Base.Rendered, generated, "the base holds no secret")
	password, err := data.Params["password"].Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, generated, password)
}

func TestProvisionFromTemplateAppliesAndDeploys(t *testing.T) {
	uc, fakes := newCreateTest(t)

	created := provision(t, uc, fakes)

	assert.True(t, fakes.envVars.applied)
	assert.Equal(t, 5432, fakes.routing.req.RoutingSettings.Port)
	assert.Equal(t, []*entity.Deployment{created.deployment}, fakes.apps.persisted.UpsertingDeployments)
	assert.Equal(t, []*entity.Task{created.deploymentTask}, fakes.apps.persisted.UpsertingTasks)
	assert.Equal(t, "postgres:18.6-alpine3.24", created.deployment.Settings.ImageSource.Image)
	assert.Equal(t, base.DeploymentTriggerSourceAPI, created.deployment.Trigger.Source)
	assert.Equal(t, "user-1", created.deployment.Trigger.SourceID)
}

func TestProvisionFromTemplateAuditsWithoutSecrets(t *testing.T) {
	uc, fakes := newCreateTest(t)

	provision(t, uc, fakes)

	assert.Len(t, fakes.audit.entries, 1)
	entry := fakes.audit.entries[0]
	assert.Equal(t, base.AuditLogTypeAppCreate, entry.Type)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Contains(t, entry.Detail, `"template":"pg"`)
	assert.Contains(t, entry.Detail, `"revision":"0123456789abcdef0123456789abcdef01234567"`)
	assert.NotContains(t, entry.Detail, fakes.templates.resp.Result.Params["password"].Text())
}

func TestProvisionFromTemplateReturnsWhatItCreatedWhenItFails(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.routing.err = errTestRouting

	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		fakes.templates.resp)

	assert.ErrorIs(t, err, errTestRouting)
	assert.Equal(t, "svc-1", created.app.ServiceID, "the caller needs the service to remove it")
}

func TestCreateAppFromTemplateRefusesBeforeProvisioning(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.templates.err = hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid)

	_, err := uc.CreateAppFromTemplate(context.Background(), testAuth(), testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
	assert.False(t, fakes.provision.called)
}

func TestCreateAppFromTemplateRefusesATemplateThatDeploysNothing(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.templates.resp.Result.Doc.Deployment = nil

	_, err := uc.CreateAppFromTemplate(context.Background(), testAuth(), testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	assert.False(t, fakes.provision.called)
}

func TestTransformAppTemplateBinding(t *testing.T) {
	appliedAt := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	resp := apptemplatedto.TransformAppTemplateBinding(&entity.AppTemplateSettings{
		Source: "official", Template: "pg", Title: "PG", Version: "18", Variant: "alpine",
		Params: map[string]*entity.AppTemplateParam{"password": {Secret: entity.NewEncryptedField("x")}},
		Base:   entity.AppTemplateBase{Revision: "abc", Release: "18.6", AppliedAt: appliedAt},
	})
	assert.Equal(t, &apptemplatedto.AppTemplateBindingResp{
		Source: "official", Template: "pg", Title: "PG", Version: "18", Release: "18.6",
		Variant: "alpine", Revision: "abc", AppliedAt: appliedAt,
	}, resp, "parameters are not part of the response")
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/`
Expected: FAIL - `undefined: createdFromTemplate`, `unknown field appProvisionService`.

- [ ] **Step 3: Write the DTOs**

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/create.go`:

```go
package apptemplatedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	appNameMaxLen     = 100
	choiceNameMaxLen  = 32
	templateNameMinLen = 1
)

type CreateAppFromTemplateReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`

	Name     string `json:"name"`
	Template string `json:"template"`
	// Version and Variant are empty for the template's defaults.
	Version string `json:"version"`
	Variant string `json:"variant"`
	// Params are JSON values of each parameter's type; a size is a string such as "1GB".
	Params map[string]any `json:"params"`
}

func NewCreateAppFromTemplateReq() *CreateAppFromTemplateReq {
	return &CreateAppFromTemplateReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *CreateAppFromTemplateReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Template = strings.TrimSpace(req.Template)
	req.Version = strings.TrimSpace(req.Version)
	req.Variant = strings.TrimSpace(req.Variant)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppFromTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 6) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, appNameMaxLen, "name")...)
	validators = append(validators,
		basedto.ValidateStr(&req.Template, true, templateNameMinLen, templateNameMaxLen, "template")...)
	validators = append(validators, basedto.ValidateStr(&req.Version, false, 1, choiceNameMaxLen, "version")...)
	validators = append(validators, basedto.ValidateStr(&req.Variant, false, 1, choiceNameMaxLen, "variant")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAppFromTemplateResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *CreateAppFromTemplateDataResp `json:"data"`
}

type CreateAppFromTemplateDataResp struct {
	App *basedto.ObjectIDResp `json:"app"`
	// Deployment is the app's first deployment, already queued: it pulls the
	// template's image.
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}
```

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/binding_get.go`:

```go
package apptemplatedto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetAppTemplateBindingReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppTemplateBindingReq() *GetAppTemplateBindingReq {
	return &GetAppTemplateBindingReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateBindingReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppTemplateBindingResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppTemplateBindingResp `json:"data"`
}

// AppTemplateBindingResp is what an app remembers of its template. Parameters are
// left out: they are inputs already applied to the app, where the settings screens
// show them - secrets masked - through their own endpoints.
//
// TODO: app templates phase 2 - the update status, the diff, applying an update
// and detaching. See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type AppTemplateBindingResp struct {
	Source    string    `json:"source"`
	Template  string    `json:"template"`
	Title     string    `json:"title"`
	Version   string    `json:"version"`
	Release   string    `json:"release"`
	Variant   string    `json:"variant"`
	Revision  string    `json:"revision"`
	AppliedAt time.Time `json:"appliedAt"`
}

func TransformAppTemplateBinding(settings *entity.AppTemplateSettings) *AppTemplateBindingResp {
	return &AppTemplateBindingResp{
		Source:    settings.Source,
		Template:  settings.Template,
		Title:     settings.Title,
		Version:   settings.Version,
		Release:   settings.Base.Release,
		Variant:   settings.Variant,
		Revision:  settings.Base.Revision,
		AppliedAt: settings.Base.AppliedAt,
	}
}
```

- [ ] **Step 4: Grow the usecase**

Replace `hivepaas_app/usecase/apptemplateuc/uc.go`:

```go
package apptemplateuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	settingRepo repository.SettingRepo

	appDeploymentService appdeploymentservice.Service
	appProvisionService  appprovisionservice.Service
	appRoutingService    approutingservice.Service
	appService           appservice.Service
	appTemplateService   apptemplateservice.Service
	auditService         auditservice.Service
	clusterService       clusterservice.Service
	envVarService        envvarservice.Service
	specService          specservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	settingRepo repository.SettingRepo,

	appDeploymentService appdeploymentservice.Service,
	appProvisionService appprovisionservice.Service,
	appRoutingService approutingservice.Service,
	appService appservice.Service,
	appTemplateService apptemplateservice.Service,
	auditService auditservice.Service,
	clusterService clusterservice.Service,
	envVarService envvarservice.Service,
	specService specservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		settingRepo: settingRepo,

		appDeploymentService: appDeploymentService,
		appProvisionService:  appProvisionService,
		appRoutingService:    appRoutingService,
		appService:           appService,
		appTemplateService:   appTemplateService,
		auditService:         auditService,
		clusterService:       clusterService,
		envVarService:        envVarService,
		specService:          specService,
	}
}
```

`hivepaas_app/usecase/apptemplateuc/create.go`:

```go
package apptemplateuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) CreateAppFromTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
) (resp *apptemplatedto.CreateAppFromTemplateResp, err error) {
	// Before the transaction: rendering may reach GitHub, and nothing is locked
	// while it does.
	rendered, err := uc.appTemplateService.Render(ctx, &apptemplateservice.RenderReq{
		Name:    req.Template,
		Version: req.Version,
		Variant: req.Variant,
		Params:  req.Params,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if doc := rendered.Result.Doc; doc.Deployment == nil || doc.Deployment.Source == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: the template deploys no image", req.Template)
	}

	var created *createdFromTemplate
	committed := false
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		// The swarm service is created inside the transaction but is not part of
		// it: a rolled-back app would leave it running with nothing recorded.
		if err != nil && !committed && created != nil && created.app.ServiceID != "" {
			_ = uc.clusterService.ServiceRemove(context.WithoutCancel(ctx), created.app.ServiceID,
				clusterservice.ItemRemovalRetryMax, 0)
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		created, txErr = uc.provisionFromTemplate(ctx, db, auth, req, rendered)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true

	// The task can be picked up only once its row exists, which is once the
	// transaction has committed.
	if err = uc.taskQueue.ScheduleTask(ctx, created.deploymentTask); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apptemplatedto.CreateAppFromTemplateResp{
		Data: &apptemplatedto.CreateAppFromTemplateDataResp{
			App:        &basedto.ObjectIDResp{ID: created.app.ID},
			Deployment: &basedto.ObjectIDResp{ID: created.deployment.ID},
		},
	}, nil
}

type createdFromTemplate struct {
	app            *entity.App
	deployment     *entity.Deployment
	deploymentTask *entity.Task
}

// provisionFromTemplate is everything that happens inside the transaction. It
// returns what it created even when it fails part way, so the caller can remove
// the swarm service a rolled-back app would otherwise leave behind.
func (uc *UC) provisionFromTemplate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	rendered *apptemplateservice.RenderResp,
) (*createdFromTemplate, error) {
	timeNow := timeutil.NowUTC()
	provisioned, err := uc.appProvisionService.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
		ProjectID:    req.ProjectID,
		ProjectEnvID: req.ProjectEnvID,
		Name:         req.Name,
		Status:       base.AppStatusActive,
		Configure: func(configureCtx context.Context, configureDB database.IDB, app *entity.App,
			spec *swarm.ServiceSpec) ([]*entity.Setting, error) {
			built, buildErr := uc.specService.BuildApp(configureCtx, configureDB, &specservice.BuildAppReq{
				App:     app,
				Doc:     rendered.Result.Doc,
				Spec:    spec,
				TimeNow: timeNow,
			})
			if buildErr != nil {
				return nil, hperrors.Wrap(buildErr)
			}
			binding, bindErr := newAppTemplateSetting(app, rendered, timeNow)
			if bindErr != nil {
				return nil, hperrors.Wrap(bindErr)
			}
			return append(built.Settings, binding), nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	created := &createdFromTemplate{app: provisioned.App}
	app := created.app

	if err = uc.applyEnvVars(ctx, db, app); err != nil {
		return created, hperrors.Wrap(err)
	}
	if err = uc.applyRouting(ctx, db, app); err != nil {
		return created, hperrors.Wrap(err)
	}
	if err = uc.createFirstDeployment(ctx, db, auth, created); err != nil {
		return created, hperrors.Wrap(err)
	}
	return created, uc.recordCreateFromTemplate(ctx, db, auth, app, rendered)
}

// applyEnvVars builds and applies the environment of every app in the new app's
// scope, as appcloneservice does after persisting a clone: the app kind makes
// shared variables such as HIVEPAAS_PASSWORD, which the app's own variables refer to.
func (uc *UC) applyEnvVars(ctx context.Context, db database.IDB, app *entity.App) error {
	// In a transaction: no nested transactions, and no concurrency.
	appEnvData, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, app.GetObjectScope(),
		false, nil, false, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, db, appEnvData, false, false)
	for _, applyErr := range errMap {
		return hperrors.Wrap(applyErr)
	}
	return nil
}

func (uc *UC) applyRouting(ctx context.Context, db database.IDB, app *entity.App) error {
	routingSetting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppRouting)
	if routingSetting == nil {
		return nil
	}
	routingSettings, err := routingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = uc.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:             app,
		RoutingSettings: routingSettings,
		RefObjects:      entity.NewRefObjects(),
	})
	return hperrors.Wrap(err)
}

// createFirstDeployment queues the deployment that replaces the placeholder
// service image with the template's.
func (uc *UC) createFirstDeployment(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	created *createdFromTemplate,
) error {
	app := created.app
	deploymentSetting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("the template deploys no image")
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	deployment, task, err := uc.appDeploymentService.CreateDeploymentAndTask(app, deploymentSettings)
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment.Trigger = &entity.AppDeploymentTrigger{
		Source:   base.DeploymentTriggerSourceAPI,
		SourceID: auth.User.ID,
	}
	err = uc.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingDeployments: []*entity.Deployment{deployment},
		UpsertingTasks:       []*entity.Task{task},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	created.deployment, created.deploymentTask = deployment, task
	return nil
}

// newAppTemplateSetting records the template an app was provisioned from. Secret
// parameters are stored encrypted, and the rendered base holds none of them.
func newAppTemplateSetting(
	app *entity.App,
	rendered *apptemplateservice.RenderResp,
	timeNow time.Time,
) (*entity.Setting, error) {
	result := rendered.Result
	data := &entity.AppTemplateSettings{
		Source:   rendered.Source,
		Template: rendered.Template.Metadata.Name,
		Title:    rendered.Template.Metadata.Title,
		Version:  result.Version.Name,
		Params:   make(map[string]*entity.AppTemplateParam, len(result.Params)),
		Base: entity.AppTemplateBase{
			Revision:       rendered.Revision,
			Release:        result.Version.Release,
			TemplateSHA256: rendered.Entry.File.SHA256,
			Rendered:       string(result.Base),
			RenderedSHA256: result.BaseSHA256,
			AppliedAt:      timeNow,
		},
	}
	if result.Variant != nil {
		data.Variant = result.Variant.Name
	}
	for name, value := range result.Params {
		if value.Value == nil {
			continue
		}
		if value.Param.Type == templatemodel.ParamTypeSecret {
			data.Params[name] = &entity.AppTemplateParam{Secret: entity.NewEncryptedField(value.Text())}
			continue
		}
		data.Params[name] = &entity.AppTemplateParam{Value: value.Text()}
	}

	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		Type:      base.SettingTypeAppTemplate,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentAppTemplateSettingsVersion,
		UpdateVer: 1,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err := setting.SetData(data); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}
```

`hivepaas_app/usecase/apptemplateuc/audit.go`:

```go
package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordCreateFromTemplate records the creation as app-create, like any other,
// with where the app came from. Parameter values are never recorded: some are
// secrets, and the rest are on the app itself.
func (uc *UC) recordCreateFromTemplate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	rendered *apptemplateservice.RenderResp,
) error {
	result := rendered.Result
	detail := auditdetail.New().
		Set("projectId", app.ProjectID).
		Set("envId", app.ProjectEnvID).
		Set("serviceId", app.ServiceID).
		Set("source", rendered.Source).
		Set("template", rendered.Template.Metadata.Name).
		Set("version", result.Version.Name).
		Set("release", result.Version.Release).
		Set("revision", rendered.Revision)
	if result.Variant != nil {
		detail.Set("variant", result.Variant.Name)
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppCreate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPICreate,
		Section:  "create",
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   detail.String(),
	})
	return hperrors.Wrap(err)
}
```

`hivepaas_app/usecase/apptemplateuc/binding_get.go`:

```go
package apptemplateuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) GetAppTemplateBinding(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateBindingReq,
) (*apptemplatedto.GetAppTemplateBindingResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app.ProjectEnvID != req.ProjectEnvID {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFound).WithParam("Name", req.AppID)
	}

	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, app.GetObjectScope(), base.SettingTypeAppTemplate, true,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.NewNotFound("App template")
		}
		return nil, hperrors.Wrap(err)
	}
	data, err := setting.AsAppTemplateSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apptemplatedto.GetAppTemplateBindingResp{
		Data: apptemplatedto.TransformAppTemplateBinding(data),
	}, nil
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/...`
Expected: PASS.

- [ ] **Step 6: Write the handlers and routes**

`hivepaas_app/interface/api/handler/apptemplatehandler/create.go`:

```go
package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// CreateAppFromTemplate Creates an app from an app template
// @Summary Creates an app from an app template
// @Description Provisions the app with the template's configuration and queues its first deployment.
// @Tags    app_templates
// @Produce json
// @Id      createAppFromTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   body body apptemplatedto.CreateAppFromTemplateReq true "request data"
// @Success 201 {object} apptemplatedto.CreateAppFromTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/from-template [post]
func (h *Handler) CreateAppFromTemplate(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthInEnv(ctx, base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewCreateAppFromTemplateReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.CreateAppFromTemplate(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}
```

`hivepaas_app/interface/api/handler/apptemplatehandler/binding_get.go`:

```go
package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateBinding Gets the template an app was created from
// @Summary Gets the template an app was created from
// @Description 404 for an app that was not created from a template.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplateBinding
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} apptemplatedto.GetAppTemplateBindingResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/template [get]
func (h *Handler) GetAppTemplateBinding(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateBindingReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplateBinding(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
```

In `hivepaas_app/interface/api/server/router_apps.go`, in the `// Base` block, after `appGroup.POST("", appHandler.CreateApp)`:

```go
		appGroup.POST("/from-template", s.handlerRegistry.appTemplateHandler.CreateAppFromTemplate)
```

and after the `// Configuration spec export` block:

```go
	{ // App templates
		appGroup.GET("/:appID/template", s.handlerRegistry.appTemplateHandler.GetAppTemplateBinding)
	}
```

- [ ] **Step 7: Build, generate swagger, run the suite**

Run: `go build ./... && make gen-swag && go test ./...`
Expected: PASS; `docs/openapi/swagger.json` gains the two operations.

- [ ] **Step 8: Try it end to end**

On a second backend reading the test repository, signed in as in Task 12 Step 7 (same commands, port 10077), with a volume id taken from the project's volume list:

```bash
VOLUME_ID=$(curl -sS "http://localhost:10077/_/projects/<projectID>/cluster-volumes" \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"][0]["id"])')

curl -sS -X POST "http://localhost:10077/_/projects/<projectID>/<env>/apps/from-template" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"name\":\"template-check\",\"template\":\"demo\",\"params\":{\"dataVolume\":\"$VOLUME_ID\"}}"
```

The project's own volumes are listed there, including the default volume every project gets. Expected: `201` with `data.app.id` and `data.deployment.id`. The demo image `demo:2.1.0` does not exist, so that deployment fails - which is still the check: the app exists with a volume mount, `app-kind` and `app-deployment` settings, and `GET /_/projects/<projectID>/<env>/apps/<appID>/template` returns `{"template":"demo","version":"2",...}`. Delete the app, then repeat with the real `postgres` template once Task 15 exists and confirm the database starts.

- [ ] **Step 9: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/usecase/apptemplateuc hivepaas_app/interface/api/handler/apptemplatehandler \
  hivepaas_app/interface/api/server/router_apps.go docs/openapi/swagger.json
git commit -m "feat(apptemplate): create an app from a template and read its binding"
```

---
## Task 14: `tools/apptemplate` and the developer docs

**Files:**
- Create: `tools/apptemplate/main.go`
- Test: `tools/apptemplate/main_test.go`
- Modify: `docs/DEVELOPMENT.md`

**Interfaces:**
- Consumes: `templaterepo.Load`, `Lint`, `BuildIndex`, `MarshalIndex`, `FindTemplate`, `ErrorText`, `IndexFile` (Task 6); `templaterender.Render`, `Request` (Task 5); the test repository `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo` (Task 11).
- Produces: `go run ./tools/apptemplate lint <dir>`, `index [-check] <dir>`, `render [-version v] [-variant v] [-param k=v]... <dir> <template>`, `pin <dir>`.

`tools/` is outside the lint configuration, but its tests run with `go test ./...`.

- [ ] **Step 1: Write the failing tests**

`tools/apptemplate/main_test.go`:

```go
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRepoDir = "../../hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo"

func copyRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	assert.NoError(t, os.CopyFS(dir, os.DirFS(testRepoDir)))
	return dir
}

func TestRepoFromRemote(t *testing.T) {
	for remote, want := range map[string]string{
		"https://github.com/hivepaas/app-templates.git":     "hivepaas/app-templates",
		"https://github.com/hivepaas/app-templates":         "hivepaas/app-templates",
		"https://github.com/hivepaas/app-templates/":        "hivepaas/app-templates",
		"git@github.com:hivepaas/app-templates.git":         "hivepaas/app-templates",
		"ssh://git@github.com/hivepaas/app-templates.git\n": "hivepaas/app-templates",
	} {
		repo, err := repoFromRemote(remote)
		assert.NoError(t, err, remote)
		assert.Equal(t, want, repo, remote)
	}

	_, err := repoFromRemote("https://gitlab.com/hivepaas/app-templates.git")
	assert.Error(t, err, "the official source reads templates from GitHub only")
}

func TestLintAndIndex(t *testing.T) {
	dir := copyRepo(t)
	var out bytes.Buffer

	assert.NoError(t, runLint([]string{dir}, &out))
	assert.Contains(t, out.String(), "OK: 1 template(s)")

	assert.ErrorIs(t, runIndex([]string{"-check", dir}, &out), errIndexStale, "there is no index.json yet")
	assert.NoError(t, runIndex([]string{dir}, &out))
	assert.NoError(t, runIndex([]string{"-check", dir}, &out))

	path := filepath.Join(dir, "templates", "demo.yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(path, append(content, []byte("# changed\n")...), 0o600))
	assert.ErrorIs(t, runIndex([]string{"-check", dir}, &out), errIndexStale, "a changed template changes its hash")
}

func TestLintReportsProblems(t *testing.T) {
	dir := copyRepo(t)
	path := filepath.Join(dir, "templates", "demo.yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(path,
		[]byte(strings.Replace(string(content), "databases/sql", "databases/graph", 1)), 0o600))
	var out bytes.Buffer

	err = runLint([]string{dir}, &out)

	assert.ErrorIs(t, err, errProblems)
	assert.Contains(t, out.String(), `category "databases/graph" is not in categories.yaml`)
}

func TestRender(t *testing.T) {
	dir := copyRepo(t)
	var out bytes.Buffer

	assert.NoError(t, runRender([]string{"-param", "dataVolume=vol-1", dir, "demo"}, &out))
	assert.Contains(t, out.String(), "demo:2.1.0")
	assert.Contains(t, out.String(), "source: vol-1")

	out.Reset()
	assert.NoError(t, runRender([]string{"-version", "1", "-param", "dataVolume=vol-1", dir, "demo"}, &out))
	assert.Contains(t, out.String(), "demo:1.9.3", "render includes deprecated versions, for authors")
}

func TestPin(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := copyRepo(t)
	var out bytes.Buffer
	assert.NoError(t, runIndex([]string{dir}, &out))
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "."},
		{"-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-qm", "templates"},
		{"remote", "add", "origin", "https://github.com/hivepaas/app-templates.git"},
	} {
		_, err := gitOutput(dir, args...)
		assert.NoError(t, err)
	}

	out.Reset()
	assert.NoError(t, runPin([]string{dir}, &out))

	var pin struct {
		Repo        string `json:"repo"`
		Commit      string `json:"commit"`
		IndexSHA256 string `json:"indexSha256"`
	}
	assert.NoError(t, json.Unmarshal(out.Bytes(), &pin))
	head, err := gitOutput(dir, "rev-parse", "HEAD")
	assert.NoError(t, err)
	index, err := os.ReadFile(filepath.Join(dir, "index.json"))
	assert.NoError(t, err)
	sum := sha256.Sum256(index)
	assert.Equal(t, "hivepaas/app-templates", pin.Repo)
	assert.Equal(t, strings.TrimSpace(string(head)), pin.Commit)
	assert.Equal(t, hex.EncodeToString(sum[:]), pin.IndexSHA256)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "index.json"), append(index, '\n'), 0o600))
	assert.Error(t, runPin([]string{dir}, &out), "a pin names the committed index, not a changed one")
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./tools/apptemplate/`
Expected: FAIL - `undefined: repoFromRemote`, `undefined: runLint`.

- [ ] **Step 3: Write the tool**

`tools/apptemplate/main.go`:

```go
// Command apptemplate works on a checkout of the app-templates repository: it
// lints templates, writes index.json, renders a template the way HivePaaS will,
// and prints the pin release.json carries.
//
// It is built on the packages HivePaaS renders templates with, so a template this
// tool accepts is one HivePaaS accepts.
//
// Usage:
//
//	go run ./tools/apptemplate lint   <dir>
//	go run ./tools/apptemplate index  [-check] <dir>
//	go run ./tools/apptemplate render [-version 18] [-variant alpine] [-param name=value]... <dir> <template>
//	go run ./tools/apptemplate pin    <dir>
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

var (
	errProblems   = errors.New("the repository has problems")
	errIndexStale = errors.New("index.json is out of date: run `apptemplate index`")

	githubRemotePattern = regexp.MustCompile(
		`^(?:https://github\.com/|git@github\.com:|ssh://git@github\.com/)([A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+?)(?:\.git)?/?$`)
)

func main() {
	if len(os.Args) < 2 { //nolint:mnd
		usage()
	}
	args := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "lint":
		err = runLint(args, os.Stdout)
	case "index":
		err = runIndex(args, os.Stdout)
	case "render":
		err = runRender(args, os.Stdout)
	case "pin":
		err = runPin(args, os.Stdout)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "apptemplate:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: apptemplate lint|index|render|pin [flags] <dir> ...")
	os.Exit(2) //nolint:mnd
}

func runLint(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("lint", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	repo, problems, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}
	problems = append(problems, templaterepo.Lint(repo)...)
	for _, problem := range problems {
		fmt.Fprintln(out, problem)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %d", errProblems, len(problems))
	}
	fmt.Fprintf(out, "OK: %d template(s)\n", len(repo.Templates))
	return nil
}

func runIndex(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	check := flags.Bool("check", false, "fail if index.json differs from what would be written")
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	data, err := buildIndex(dir)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, templaterepo.IndexFile)
	if *check {
		current, readErr := os.ReadFile(path)
		if readErr != nil || !bytes.Equal(current, data) {
			return errIndexStale
		}
		fmt.Fprintln(out, "index.json is up to date")
		return nil
	}
	if err = os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec
		return err
	}
	fmt.Fprintf(out, "written: %s\n", path)
	return nil
}

// buildIndex refuses a repository with problems: an index is published, and an
// index of broken templates is a store of forms that cannot be submitted.
func buildIndex(dir string) ([]byte, error) {
	repo, problems, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return nil, errors.New(templaterepo.ErrorText(err))
	}
	if problems = append(problems, templaterepo.Lint(repo)...); len(problems) > 0 {
		return nil, fmt.Errorf("%w: run `apptemplate lint` to see them", errProblems)
	}
	index, err := templaterepo.BuildIndex(repo)
	if err != nil {
		return nil, errors.New(templaterepo.ErrorText(err))
	}
	return templaterepo.MarshalIndex(index)
}

type paramFlags map[string]any

func (p paramFlags) String() string {
	pairs := make([]string, 0, len(p))
	for name, value := range p {
		pairs = append(pairs, fmt.Sprintf("%s=%v", name, value))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

func (p paramFlags) Set(value string) error {
	name, v, found := strings.Cut(value, "=")
	if !found || name == "" {
		return fmt.Errorf("-param %q: want name=value", value)
	}
	p[name] = v
	return nil
}

func runRender(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	version := flags.String("version", "", "version to render; empty for the default")
	variant := flags.String("variant", "", "variant to render; empty for the default")
	params := paramFlags{}
	flags.Var(params, "param", "a parameter as name=value; repeatable")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 2 { //nolint:mnd
		return errors.New("usage: apptemplate render [flags] <dir> <template>")
	}

	repo, _, err := templaterepo.Load(os.DirFS(flags.Arg(0)))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}
	file := repo.FindTemplate(flags.Arg(1))
	if file == nil {
		return fmt.Errorf("no template %q in %s (a template that does not load is left out: run lint)",
			flags.Arg(1), flags.Arg(0))
	}

	result, err := templaterender.Render(&templaterender.Request{
		Template:        file.Template,
		Version:         *version,
		Variant:         *variant,
		Params:          params,
		AllowDeprecated: true,
	})
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}

	fmt.Fprintf(out, "# %s %s (%s), base sha256 %s\n", file.Template.Metadata.Name, result.Version.Name,
		result.Image, result.BaseSHA256)
	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(2) //nolint:mnd
	if err = encoder.Encode(result.Doc); err != nil {
		return err
	}
	return encoder.Close()
}

type pin struct {
	Repo        string `json:"repo"`
	Commit      string `json:"commit"`
	IndexSHA256 string `json:"indexSha256"`
}

// runPin prints what release.json's templates field carries for the checkout's
// HEAD. The hash is of index.json as committed, because that is the file the
// official source fetches by commit.
func runPin(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("pin", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	status, err := gitOutput(dir, "status", "--porcelain", "--", templaterepo.IndexFile)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(status)) > 0 {
		return errors.New("index.json has uncommitted changes: the pin names the committed file")
	}
	commit, err := gitOutput(dir, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	index, err := gitOutput(dir, "show", "HEAD:"+templaterepo.IndexFile)
	if err != nil {
		return err
	}
	remote, err := gitOutput(dir, "remote", "get-url", "origin")
	if err != nil {
		return err
	}
	repo, err := repoFromRemote(string(remote))
	if err != nil {
		return err
	}

	sum := sha256.Sum256(index)
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pin{
		Repo:        repo,
		Commit:      strings.TrimSpace(string(commit)),
		IndexSHA256: hex.EncodeToString(sum[:]),
	})
}

// repoFromRemote reads owner/name from a GitHub remote. Only GitHub: the
// official source builds raw.githubusercontent.com URLs from it.
func repoFromRemote(remote string) (string, error) {
	match := githubRemotePattern.FindStringSubmatch(strings.TrimSpace(remote))
	if match == nil {
		return "", fmt.Errorf("remote %q is not a GitHub repository", strings.TrimSpace(remote))
	}
	return match[1], nil
}

func gitOutput(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func singleDir(flags *flag.FlagSet) (string, error) {
	if flags.NArg() != 1 {
		return "", fmt.Errorf("usage: apptemplate %s <dir>", flags.Name())
	}
	return flags.Arg(0), nil
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./tools/apptemplate/`
Expected: PASS.

- [ ] **Step 5: Document authoring in `docs/DEVELOPMENT.md`**

Add a row to the table under `## At a glance`, after the spec export row:

```markdown
| App templates from a local checkout | `HP_TEMPLATES_DIR=../app-templates make local-app-run` | the project Store tab, §7 |
```

Rename `## 7. Before you push` to `## 8. Before you push`, and insert this section before it:

````markdown
## 7. App templates

Templates live in their own repository,
[hivepaas/app-templates](https://github.com/hivepaas/app-templates), cloned next to
this one as `../app-templates`. The design is in
[specs/2026-09-17-app-templates-design.md](superpowers/specs/2026-09-17-app-templates-design.md).

An installation only reads templates pinned by the signed release info. To try a
template before it is published, point a development backend at the checkout:

```bash
HP_TEMPLATES_DIR=../app-templates make local-app-run
```

`HP_TEMPLATES_DIR` is read with no signature and no hashes, re-read on every
request, and ignored - with a warning in the log - anywhere `env` is not
`development`. The store shows the checkout's templates with source `local`.

### The tool

```bash
go run ./tools/apptemplate lint   ../app-templates
go run ./tools/apptemplate render -param dataVolume=<volume-id> ../app-templates postgres
go run ./tools/apptemplate index  ../app-templates           # rewrite index.json
go run ./tools/apptemplate index  -check ../app-templates    # what CI runs
go run ./tools/apptemplate pin    ../app-templates           # on the merged commit
```

The tool renders with the packages HivePaaS renders with, so a template that
lints here renders there.

### Publishing

1. A pull request to `app-templates`; its CI runs `lint` and `index -check`.
2. On the merged commit, `go run ./tools/apptemplate pin ../app-templates`.
3. Put the printed object into `release.json` as `templates` under `beta`,
   `make release-sign`, and commit both files to the `release` branch.
4. Check the store on a beta installation, then copy the pin to `stable` and
   sign again.

A template reaches installations without a HivePaaS release, but never without
the offline signature: a template decides which images run.
````

- [ ] **Step 6: Commit**

```bash
git add tools/apptemplate docs/DEVELOPMENT.md
git commit -m "feat(apptemplate): lint, index, render and pin tool, with authoring docs"
```

---
## Task 15: The `app-templates` repository

This task works in the other repository, `/Users/tnt/go/src/github.com/hivepaas/app-templates` (it holds only `README.md` and `LICENSE` today), and uses `tools/apptemplate` from this one.

**Files (all in `../app-templates`):**
- Modify: `README.md`
- Create: `categories.yaml`, `tags.yaml`
- Create: `templates/postgres.yaml`, `templates/mariadb.yaml`
- Create: `icons/postgres.svg`, `icons/mariadb.svg`
- Create: `.github/workflows/lint.yml`
- Create: `index.json` (generated)

**Interfaces:**
- Consumes: `go run ./tools/apptemplate lint|index|render` (Task 14); the format (Tasks 1-6); `HP_TEMPLATES_DIR` (Task 11); `POST .../apps/from-template` (Task 13).
- Produces: a repository `apptemplate lint` accepts, with an `index.json` `apptemplate index -check` accepts - what the release's `templates` pin will point at.

The image tags below were current on Docker Hub when this plan was written (postgres 18.6, 17.11 and 16.15; mariadb 11.8.9 and 11.4.13). Redis is left out on purpose: the `cache` kind produces no `HIVEPAAS_PASSWORD` for an app to read yet (`app_system_env.go` carries a TODO for it), so a redis template would have to put its password into a plain environment variable.

- [ ] **Step 1: Write the vocabularies**

`categories.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: TemplateCategories
categories:
  - id: databases
    title: Databases
    children:
      - {id: sql, title: SQL}
      - {id: nosql, title: NoSQL}
      - {id: cache, title: Cache & Queues}
  - id: webapps
    title: Web Apps
    children:
      - {id: dev-tools, title: Developer Tools}
      - {id: productivity, title: Productivity}
      - {id: cms, title: CMS & Blogging}
  - id: observability
    title: Observability
    children:
      - {id: monitoring, title: Monitoring}
      - {id: analytics, title: Analytics}
  - id: storage
    title: Storage
    children:
      - {id: object-storage, title: Object Storage}
```

`tags.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: TemplateTags
tags:
  - {id: sql, title: SQL}
  - {id: relational, title: Relational}
  - {id: open-source, title: Open Source}
  - {id: mysql-compatible, title: MySQL Compatible}
```

- [ ] **Step 2: Write the postgres template**

`templates/postgres.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate

metadata:
  name: postgres
  title: PostgreSQL
  tagline: The world's most advanced open source relational database
  description: |
    PostgreSQL is a powerful, open source object-relational database with a
    strong reputation for reliability, feature robustness and performance.

    **This template sets up**

    - a database, a user and a generated password
    - a persistent data volume
    - a health check with `pg_isready`
    - a memory limit you choose

    Other apps in the same environment reach it by the app's name on port 5432,
    with the credentials HivePaaS shares as `HIVEPAAS_USER`, `HIVEPAAS_PASSWORD`
    and `HIVEPAAS_DATABASE_NAME`.
  categories: [databases/sql]
  tags: [sql, relational, open-source]
  aliases: [pg, psql, postgresql]
  icon: icons/postgres.svg
  links:
    website: https://www.postgresql.org
    documentation: https://www.postgresql.org/docs/
    source: https://github.com/postgres/postgres
  license: PostgreSQL
  requires:
    versionCode: v000001

parameters:
  - name: dbName
    title: Database name
    type: string
    default: app
    pattern: '^[a-z_][a-z0-9_]{0,62}$'
  - name: username
    title: Username
    type: string
    default: app
    pattern: '^[a-z_][a-z0-9_]{0,62}$'
  - name: password
    title: Password
    description: Leave empty to generate one.
    type: secret
    minLength: 12
    generate: {length: 32}
  - name: memoryLimit
    title: Memory limit
    type: size
    default: 1GB
    min: 256MB
  - name: dataVolume
    title: Data volume
    description: The volume the database files are kept on.
    type: volume

variants:
  - name: alpine
    title: Alpine
    description: A smaller image, based on musl libc
    default: true
  - name: debian
    title: Debian
    description: Based on glibc, for extensions built against it

versions:
  - name: "18"
    release: "18.6"
    default: true
    images:
      alpine: postgres:18.6-alpine3.24
      debian: postgres:18.6-trixie
  - name: "17"
    release: "17.11"
    images:
      alpine: postgres:17.11-alpine3.24
      debian: postgres:17.11-trixie
    override:
      app:
        deployment:
          storage:
            mounts:
              # Before 18, the image keeps its data directly in /var/lib/postgresql/data.
              /var/lib/postgresql: null
              /var/lib/postgresql/data: {type: volume, source: "${{ params.dataVolume }}"}
  - name: "16"
    release: "16.15"
    images:
      alpine: postgres:16.15-alpine3.24
      debian: postgres:16.15-trixie
    override:
      app:
        deployment:
          storage:
            mounts:
              /var/lib/postgresql: null
              /var/lib/postgresql/data: {type: volume, source: "${{ params.dataVolume }}"}

app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        # From 18 on, the image declares its volume at /var/lib/postgresql and keeps
        # each major's data in its own directory under it.
        /var/lib/postgresql: {type: volume, source: "${{ params.dataVolume }}"}
    container:
      healthcheck:
        enabled: true
        mode: CMD-SHELL
        command: pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"
        interval: 10s
        timeout: 5s
        startPeriod: 30s
        retries: 5
    resources:
      limits: {memory: "${{ params.memoryLimit }}"}
  settings:
    kind:
      category: database
      engine: postgres
      version: "${{ version.name }}"
      database:
        dbName: "${{ params.dbName }}"
        username: "${{ params.username }}"
        password: "${{ params.password }}"
    envVars:
      data:
        - {k: POSTGRES_DB, v: "${HIVEPAAS_DATABASE_NAME}"}
        - {k: POSTGRES_USER, v: "${HIVEPAAS_USER}"}
        - {k: POSTGRES_PASSWORD, v: "${HIVEPAAS_PASSWORD}"}
    routing:
      port: 5432
```

- [ ] **Step 3: Write the mariadb template**

`templates/mariadb.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate

metadata:
  name: mariadb
  title: MariaDB
  tagline: A community-developed fork of MySQL, fast and open source
  description: |
    MariaDB Server is one of the most popular open source relational databases,
    made by the original developers of MySQL and compatible with it.

    **This template sets up**

    - a database, a user, and generated user and root passwords
    - a persistent data volume
    - a health check with the image's `healthcheck.sh`
    - a memory limit you choose

    Other apps in the same environment reach it by the app's name on port 3306,
    with the credentials HivePaaS shares as `HIVEPAAS_USER`, `HIVEPAAS_PASSWORD`
    and `HIVEPAAS_DATABASE_NAME`. The root password stays with this app.
  categories: [databases/sql]
  tags: [sql, relational, open-source, mysql-compatible]
  aliases: [mysql]
  icon: icons/mariadb.svg
  links:
    website: https://mariadb.org
    documentation: https://mariadb.com/kb/en/documentation/
    source: https://github.com/MariaDB/server
  license: GPL-2.0
  requires:
    versionCode: v000001

parameters:
  - name: dbName
    title: Database name
    type: string
    default: app
    pattern: '^[A-Za-z_][A-Za-z0-9_]{0,63}$'
  - name: username
    title: Username
    type: string
    default: app
    pattern: '^[A-Za-z_][A-Za-z0-9_]{0,79}$'
  - name: password
    title: Password
    description: Leave empty to generate one.
    type: secret
    minLength: 12
    generate: {length: 32}
  - name: rootPassword
    title: Root password
    description: Leave empty to generate one.
    type: secret
    minLength: 12
    generate: {length: 32}
  - name: memoryLimit
    title: Memory limit
    type: size
    default: 1GB
    min: 256MB
  - name: dataVolume
    title: Data volume
    description: The volume the database files are kept on.
    type: volume

versions:
  - name: "11.8"
    release: "11.8.9"
    default: true
    image: mariadb:11.8.9-noble
  - name: "11.4"
    release: "11.4.13"
    image: mariadb:11.4.13-noble

app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        /var/lib/mysql: {type: volume, source: "${{ params.dataVolume }}"}
    container:
      healthcheck:
        enabled: true
        mode: CMD-SHELL
        command: healthcheck.sh --connect --innodb_initialized
        interval: 10s
        timeout: 5s
        startPeriod: 30s
        retries: 5
    resources:
      limits: {memory: "${{ params.memoryLimit }}"}
  settings:
    kind:
      category: database
      engine: mariadb
      version: "${{ version.name }}"
      database:
        dbName: "${{ params.dbName }}"
        username: "${{ params.username }}"
        password: "${{ params.password }}"
        rootPassword: "${{ params.rootPassword }}"
    envVars:
      data:
        - {k: MARIADB_DATABASE, v: "${HIVEPAAS_DATABASE_NAME}"}
        - {k: MARIADB_USER, v: "${HIVEPAAS_USER}"}
        - {k: MARIADB_PASSWORD, v: "${HIVEPAAS_PASSWORD}"}
        - {k: MARIADB_ROOT_PASSWORD, v: "${HIVEPAAS_ROOT_PASSWORD}"}
    routing:
      port: 3306
```

- [ ] **Step 4: Draw the icons**

Plain database glyphs, not the projects' logos - those are trademarks, and a logo in a store implies an endorsement nobody gave.

`icons/postgres.svg`:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="PostgreSQL">
  <rect width="64" height="64" rx="14" fill="#336791"/>
  <g fill="none" stroke="#ffffff" stroke-width="3">
    <ellipse cx="32" cy="18" rx="16" ry="6"/>
    <path d="M16 18v28c0 3.3 7.2 6 16 6s16-2.7 16-6V18"/>
    <path d="M16 32c0 3.3 7.2 6 16 6s16-2.7 16-6"/>
  </g>
</svg>
```

`icons/mariadb.svg`:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="MariaDB">
  <rect width="64" height="64" rx="14" fill="#003545"/>
  <g fill="none" stroke="#c0765a" stroke-width="3">
    <ellipse cx="32" cy="18" rx="16" ry="6"/>
    <path d="M16 18v28c0 3.3 7.2 6 16 6s16-2.7 16-6V18"/>
    <path d="M16 32c0 3.3 7.2 6 16 6s16-2.7 16-6"/>
  </g>
</svg>
```

- [ ] **Step 5: Lint, render and index**

From the hivepaas repository:

```bash
go run ./tools/apptemplate lint ../app-templates
go run ./tools/apptemplate render -param dataVolume=vol-1 ../app-templates postgres | head -20
go run ./tools/apptemplate render -version 17 -variant debian -param dataVolume=vol-1 ../app-templates postgres \
  | grep -A2 'mounts:'
go run ./tools/apptemplate index ../app-templates
go run ./tools/apptemplate index -check ../app-templates
```

Expected: `OK: 2 template(s)`; the first render shows `postgres:18.6-alpine3.24` and a mount at `/var/lib/postgresql`; the second shows `postgres:17.11-trixie` and a mount at `/var/lib/postgresql/data`; `written: ../app-templates/index.json`; `index.json is up to date`.

- [ ] **Step 6: Check every image exists**

```bash
for image in postgres:18.6-alpine3.24 postgres:18.6-trixie postgres:17.11-alpine3.24 postgres:17.11-trixie \
  postgres:16.15-alpine3.24 postgres:16.15-trixie mariadb:11.8.9-noble mariadb:11.4.13-noble; do
  docker manifest inspect "$image" >/dev/null && echo "ok $image" || echo "MISSING $image"
done
```

Expected: eight `ok` lines. A `MISSING` tag was replaced upstream: take the current patch tag from Docker Hub, update `release` and the image together, and run Step 5 again.

- [ ] **Step 7: Provision both for real**

Run a second backend on its own port against the checkout, and sign in as in Task 12 Step 7:

```bash
HP_HTTP_SERVER_PORT=10077 HP_RUN_MODE=app+worker HP_TEMPLATES_DIR=../app-templates make local-app-run &
```

(`app+worker`, because the first deployment is a task the worker runs.) With `TOKEN` and `VOLUME_ID` as in Task 13 Step 8:

```bash
for template in postgres mariadb; do
  curl -sS -X POST "http://localhost:10077/_/projects/<projectID>/<env>/apps/from-template" \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d "{\"name\":\"$template-check\",\"template\":\"$template\",\"params\":{\"dataVolume\":\"$VOLUME_ID\"}}"
done
```

Expected: both `201`. Within a couple of minutes both deployments succeed and `docker ps` lists a container for each whose status ends in `(healthy)`. The credentials reached the database:

```bash
docker exec "$(docker ps -q --filter name=postgres-check)" \
  sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "select 1"'
docker exec "$(docker ps -q --filter name=mariadb-check)" \
  sh -c 'mariadb -u"$MARIADB_USER" -p"$MARIADB_PASSWORD" "$MARIADB_DATABASE" -Nse "select 1"'
```

Expected: `1` twice. Delete both apps and stop the second backend.

- [ ] **Step 8: Write the README**

`README.md`:

````markdown
# app-templates

The app templates HivePaaS offers in its store: common apps - databases first -
provisioned in one step, with configuration a maintainer has already got right.

An installation reads this repository only through its signed release info, which
pins a commit and the sha256 of `index.json`; the index pins the sha256 of every
template and icon. A change here reaches nobody until it is published (see
Publishing).

## Layout

```
categories.yaml     the category tree
tags.yaml           the tag vocabulary
templates/<name>.yaml
icons/<name>.svg    or .png, at most 256 KB
index.json          generated by the tool - never edit it
```

## A template

A template is one YAML file describing one app. The two in `templates/` are the
reference; the rules are:

- **metadata** - `name` equals the file name; `tagline` at most 80 characters;
  `description` is markdown and is rendered without raw HTML; 1 to 3
  `categories` written `parent/child` from `categories.yaml`; at most 8 `tags`
  from `tags.yaml`; `aliases` are search terms that are never shown;
  `requires.versionCode` is the oldest HivePaaS that can provision it.
- **parameters** - what the create form asks for. Types: `string`, `secret`,
  `int`, `size` (`512MB`), `bool`, `select`, `volume`. A `secret` has no
  `default`; give it `generate` to create one when left empty. A `volume` is a
  cluster volume the user picks.
- **versions** - one per major line, each pinned to an exact image. `release` is
  the exact version shown to users. Moving tags such as `latest` or `17` are
  refused. `deprecated` hides a version from new apps.
- **variants** - distributions of the same software, such as alpine and debian.
  With variants, each version lists `images` by variant; without, it has `image`.
- **override.app** - what a version changes in `app`, as a JSON merge patch:
  maps merge, `null` deletes a key, lists are replaced.
- **app** - the app itself, in the shape a HivePaaS configuration spec uses.
  Supported: `deployment.source` (an image), `deployment.storage.mounts`
  (volumes), `deployment.container.healthcheck`, `deployment.resources`, and
  `settings.kind`, `settings.envVars`, `settings.routing.port`.

Placeholders are `${{ params.<name> }}`, `${{ version.name }}`,
`${{ version.release }}`, `${{ version.vars.<name> }}`, `${{ variant.name }}` and
`${{ image }}`; write `$${{` for a literal `${{`. `${HIVEPAAS_PASSWORD}` and other
`${...}` references are HivePaaS's own environment variables and pass through
untouched.

## Working on templates

The tool lives in the [hivepaas](https://github.com/hivepaas/hivepaas) repository,
so run it from a checkout of that next to this one:

```bash
go run ./tools/apptemplate lint   ../app-templates
go run ./tools/apptemplate render -param dataVolume=<volume-id> ../app-templates postgres
go run ./tools/apptemplate index  ../app-templates
```

To try a template in a local HivePaaS before publishing it, start the backend with
`HP_TEMPLATES_DIR=../app-templates` (development environment only).

## Publishing

1. Open a pull request. CI lints every template and checks `index.json` is current.
2. After merging, run `go run ./tools/apptemplate pin ../app-templates` on the merged
   commit.
3. The HivePaaS release owner puts that pin into `release.json`, signs it offline,
   and publishes it - to beta first, then stable.
````

- [ ] **Step 9: Add CI**

`.github/workflows/lint.yml`:

```yaml
name: lint

on:
  pull_request:
  push:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          path: app-templates

      - uses: actions/checkout@v4
        with:
          repository: hivepaas/hivepaas
          # The HivePaaS commit whose parser checks these templates. Pinned, so a
          # template is checked by the code that will read it; move it when HivePaaS
          # gains a block or a parameter type.
          ref: HIVEPAAS_COMMIT_SHA
          path: hivepaas

      - uses: actions/setup-go@v5
        with:
          go-version-file: hivepaas/go.mod
          cache-dependency-path: hivepaas/go.sum

      - name: Lint templates
        working-directory: hivepaas
        run: go run ./tools/apptemplate lint ../app-templates

      - name: Check index.json is current
        working-directory: hivepaas
        run: go run ./tools/apptemplate index -check ../app-templates
```

`HIVEPAAS_COMMIT_SHA` is replaced in Step 10, once the hivepaas commit that contains `tools/apptemplate` exists.

- [ ] **Step 10: Pin CI and commit**

After Tasks 1-14 are committed in hivepaas, put that commit into the workflow and commit this repository:

```bash
cd ../app-templates
sed -i '' "s/HIVEPAAS_COMMIT_SHA/$(git -C ../hivepaas rev-parse HEAD)/" .github/workflows/lint.yml
(cd ../hivepaas && go run ./tools/apptemplate index -check ../app-templates)

git add README.md categories.yaml tags.yaml templates icons index.json .github/workflows/lint.yml
git commit -m "Add postgres and mariadb templates"
```
 Push when the hivepaas commit named in the workflow is pushed too, or CI cannot check it out.

---

## After all tasks

- [ ] `go build ./... && make lint-local && go test ./...` in hivepaas - all green.
- [ ] `grep -rn "TODO: app templates\|TODO: spec import" hivepaas_app` lists the deferred work at the places §12 of the spec names: parameter types, blocks, variant override, update policy, the binding response, mount source mapping, spec import.
- [ ] Publishing (`pin`, `release.json`, `make release-sign`) is a release step for the owner of the signing keys, not part of this plan.
