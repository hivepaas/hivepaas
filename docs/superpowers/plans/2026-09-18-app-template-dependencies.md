# App Template Dependencies Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a template name other templates in the same repository as dependencies, so that creating a web app from a template also creates its database and wires the credentials in.

**Architecture:** The template format gains a `dependencies` block and two placeholders, `${{ deps.<name>.key }}` and `${{ deps.<name>.ref.<VAR> }}`, which render to an app key and to an ordinary `${<key>.<VAR>}` environment reference. The service renders every app of a request before any exists - dependencies first, since the owner's placeholders need what they share - and the use case provisions them in one transaction, dependencies first, with ids chosen up front so each binding can name the others. The linter checks a template against the templates it names.

**Tech Stack:** Go 1.27, gin, bun/PostgreSQL, Docker Swarm, gopkg.in/yaml.v3, testify.

**Spec:** [docs/superpowers/specs/2026-09-17-app-template-dependencies-design.md](../specs/2026-09-17-app-template-dependencies-design.md)

## Global Constraints

- Dependencies go **one level** deep: a template named as a dependency may not declare dependencies.
- **At most 3** dependencies per template (`templatemodel.MaxDependencies = 3`), names unique.
- A dependency names a template **in the same repository**, loaded from the same revision.
- A dependency is **always created**; attaching an existing app is out of scope.
- Deleting an app **never deletes** its dependencies.
- Dependencies are resolved **at creation only**.
- A dependency app is named `<app name>-<dependency name>`; its key is `projecthelper.CalcAppKey` of that name.
- `${{ deps.<name>.ref.<VAR> }}` renders to `${<key>.<VAR>}`, and `<VAR>` must be a variable the dependency's kind shares.
- Template files stay strictly decoded. `index.json` accepts fields it does not know.
- The dashboard is out of scope: the store UI does not exist yet. The wire additions of Task 9 are what its plan builds on.
- Checks before a task is done: `go build ./...`, `golangci-lint run ./...` over the **whole** repository (120-character lines, US spelling), `go test ./...`. `make gen-swag` when a DTO changes.
- Commits end with `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`.
- Work on branch `feat/app-template-dependencies`, created from `feat/postgres-extension-templates` (it carries the storage kind and the spec). The templates work goes on `feat/webapp-templates` in `../app-templates`, created from `feat/database-templates`.

---

## File Structure

**hivepaas**

| File | Responsibility |
|---|---|
| `hivepaas_app/service/apptemplateservice/templatemodel/index.go` | Tolerant index decoding; `IndexEntry.Dependencies`, `IndexDependency` |
| `hivepaas_app/service/apptemplateservice/templatemodel/dependency.go` (new) | `Dependency`, its validation, `FindDependency`, `AskedParams`, `DependencyAppName` |
| `hivepaas_app/service/apptemplateservice/templatemodel/template.go` | `Template.Dependencies` field |
| `hivepaas_app/service/apptemplateservice/templatemodel/validate.go` | Calls the dependency validation |
| `hivepaas_app/base/env_var.go` | `AppCommonSharedEnvVars`, `AppKindSharedEnvVars` |
| `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go` | Holds `kindEnvVars` and `base` together |
| `hivepaas_app/service/apptemplateservice/templaterender/deps.go` (new) | `DepBinding`, `deps.*` resolution, `DependencyParams`, `SharedVarsOf` |
| `hivepaas_app/service/apptemplateservice/templaterender/render.go` | `Request.ResolvedParams`, `Request.Deps`, resolver wiring |
| `hivepaas_app/service/apptemplateservice/templaterepo/lint_dependencies.go` (new) | Repository-level dependency checks and bound renders |
| `hivepaas_app/service/apptemplateservice/templaterepo/lint.go` | Calls them |
| `hivepaas_app/service/apptemplateservice/templaterepo/index.go` | Fills `IndexEntry.Dependencies` |
| `hivepaas_app/service/apptemplateservice/types.go` | Dependencies on `TemplateResp`, `RenderReq`, `RenderResp` |
| `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go` | Loads and renders dependencies |
| `hivepaas_app/service/appprovisionservice/types.go`, `.../appprovisionserviceimpl/provision.go` | `ProvisionAppReq.AppID` |
| `hivepaas_app/entity/setting_app_template.go` | `Dependencies`, `CreatedForAppID` on the binding |
| `hivepaas_app/usecase/apptemplateuc/app_create_from_template.go` | Plans, provisions, cleans up and schedules every app of a request |
| `hivepaas_app/usecase/apptemplateuc/audit.go` | Records the links |
| `hivepaas_app/usecase/apptemplateuc/apptemplatedto/*.go` | Wire format |

**app-templates**: `templates/wordpress.yaml`, `icons/wordpress.svg`, `tags.yaml`, `README.md`, `index.json`.

---

### Task 1: The index accepts fields it does not know

Every installation reads the index its channel pins, whatever its own HivePaaS version, so a new `IndexEntry` field - `license` already, `dependencies` next - takes the store down on every older installation while `DecodeIndex` refuses unknown fields. App templates are unreleased, so nothing is broken yet; this fixes it before anything ships.

**Files:**
- Modify: `hivepaas_app/service/apptemplateservice/templatemodel/index.go` (`DecodeIndex`)
- Test: `hivepaas_app/service/apptemplateservice/templatemodel/template_test.go` (`TestDecodeIndexRefusesAnotherKindAndUnknownFields`)

**Interfaces:**
- Produces: `templatemodel.DecodeIndex(data []byte) (*Index, error)` - same signature, now tolerant.

- [ ] **Step 1: Branch**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
git checkout feat/postgres-extension-templates
git checkout -b feat/app-template-dependencies
```

- [ ] **Step 2: Change the test to the new contract**

In `template_test.go`, replace `TestDecodeIndexRefusesAnotherKindAndUnknownFields` with:

```go
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
			"versions": [], "requires": {"versionCode": "v000001"}}]
	}`))

	assert.NoError(t, err)
	assert.Equal(t, "Demo", index.FindTemplate("demo").Title)
}
```

- [ ] **Step 3: Run it and watch it fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ -run TestDecodeIndex -v`
Expected: FAIL in `TestDecodeIndexAcceptsFieldsItDoesNotKnow` with `json: unknown field "addedLater"`.

- [ ] **Step 4: Make `DecodeIndex` tolerant**

Replace the function in `index.go`:

```go
// DecodeIndex parses index.json. Unlike a template file, it accepts fields it
// does not know.
//
// Every installation reads the index its channel pins, whatever HivePaaS version
// it runs, so an index gains fields older installations have never heard of.
// Refusing them would take the store down on every one of those installations
// the moment a newer index is pinned. That is safe here and would not be for a
// template: the index only lists, and nothing is provisioned from what an older
// HivePaaS skipped in it. Template files stay strict, and requires.versionCode is
// what keeps an older HivePaaS from provisioning a template it cannot read.
func DecodeIndex(data []byte) (*Index, error) {
	index := &Index{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(index); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("index.json: %s", err.Error())
	}
	if err := checkHeader("index.json", index.APIVersion, index.Kind, KindTemplateIndex); err != nil {
		return nil, err
	}
	return index, nil
}
```

- [ ] **Step 5: Run the package**

Run: `go test ./hivepaas_app/service/apptemplateservice/...`
Expected: `ok` for every package.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/apptemplateservice/templatemodel/index.go \
  hivepaas_app/service/apptemplateservice/templatemodel/template_test.go
git commit -m "$(cat <<'EOF'
fix(apptemplate): read an index that has fields this version does not know

Every installation reads the index its channel pins, whatever version it
runs, so a field added to the index - license already - would take the
store down on every older installation. The index only lists; nothing is
provisioned from a field an older HivePaaS skips in it. Template files stay
strict.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: The `dependencies` block in the template model

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/dependency.go`
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/dependency_test.go`
- Modify: `hivepaas_app/service/apptemplateservice/templatemodel/template.go` (`Template`)
- Modify: `hivepaas_app/service/apptemplateservice/templatemodel/validate.go` (`Template.Validate`)
- Modify: `hivepaas_app/service/apptemplateservice/templatemodel/index.go` (`IndexEntry`)

**Interfaces:**
- Produces:
  - `type Dependency struct { Name, Title, Template, Version, Variant string; Params map[string]any }` (yaml keys `name`, `title`, `template`, `version`, `variant`, `params`)
  - `const MaxDependencies = 3`
  - `func (t *Template) FindDependency(name string) *Dependency`
  - `func (d *Dependency) AskedParams(depTemplate *Template) []*Parameter`
  - `func DependencyAppName(appName, dependencyName string) string`
  - `Template.Dependencies []*Dependency`
  - `IndexEntry.Dependencies []*IndexDependency`, `type IndexDependency struct { Name, Title, Template string }` (json `name`, `title`, `template`)

- [ ] **Step 1: Write the failing tests**

Create `dependency_test.go`:

```go
package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func withDependency(t *testing.T) *Template {
	t.Helper()
	tmpl := mustDecode(t)
	tmpl.Dependencies = []*Dependency{{
		Name: "db", Title: "Database", Template: "mysql", Version: "8.4",
		Params: map[string]any{"dbName": "wordpress"},
	}}
	return tmpl
}

func TestValidateAcceptsADependency(t *testing.T) {
	assert.NoError(t, withDependency(t).Validate("demo"))
}

func TestValidateRefusesABadDependency(t *testing.T) {
	cases := map[string]struct {
		mutate func(tmpl *Template)
		want   string
	}{
		"too many": {func(tm *Template) {
			for _, name := range []string{"cache", "search", "queue"} {
				tm.Dependencies = append(tm.Dependencies, &Dependency{Name: name, Title: name, Template: "redis"})
			}
		}, "at most 3"},
		"uppercase name":   {func(tm *Template) { tm.Dependencies[0].Name = "DB" }, "lowercase letters and digits"},
		"dashed name":      {func(tm *Template) { tm.Dependencies[0].Name = "my-db" }, "lowercase letters and digits"},
		"no title":         {func(tm *Template) { tm.Dependencies[0].Title = "" }, "title is required"},
		"bad template":     {func(tm *Template) { tm.Dependencies[0].Template = "My SQL" }, "is not a template name"},
		"itself":           {func(tm *Template) { tm.Dependencies[0].Template = "demo" }, "cannot depend on itself"},
		"bad version":      {func(tm *Template) { tm.Dependencies[0].Version = "8.4 lts" }, `version "8.4 lts" is invalid`},
		"bad param name":   {func(tm *Template) { tm.Dependencies[0].Params["db name"] = "x" }, "is not a parameter name"},
		"declared twice": {func(tm *Template) {
			tm.Dependencies = append(tm.Dependencies, &Dependency{Name: "db", Title: "Again", Template: "postgres"})
		}, "declared twice"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tmpl := withDependency(t)
			tc.mutate(tmpl)
			err := tmpl.Validate("demo")
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
			assert.Contains(t, errorDetail(t, err), tc.want)
		})
	}
}

func TestAskedParams(t *testing.T) {
	database := &Template{Parameters: []*Parameter{
		{Name: "dbName", Type: ParamTypeString, Default: "app"},
		{Name: "username", Type: ParamTypeString},
		{Name: "password", Type: ParamTypeSecret, Generate: &Generate{Length: 32}},
		{Name: "note", Type: ParamTypeString, Optional: true},
		{Name: "blank", Type: ParamTypeString, Default: ""},
		{Name: "dataVolume", Type: ParamTypeVolume},
	}}
	dep := &Dependency{Name: "db", Params: map[string]any{"username": "wordpress"}}

	var names []string
	for _, param := range dep.AskedParams(database) {
		names = append(names, param.Name)
	}

	assert.Equal(t, []string{"blank", "dataVolume"}, names,
		"fixed, defaulted, optional and generated parameters are not asked; an empty default is no default")
}

func TestDependencyAppName(t *testing.T) {
	assert.Equal(t, "blog-db", DependencyAppName("blog", "db"))
}
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ -run 'Dependenc|AskedParams' -v`
Expected: FAIL to compile: `undefined: Dependency`.

- [ ] **Step 3: Write `dependency.go`**

```go
package templatemodel

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
)

// MaxDependencies caps what one template creates besides its own app. One
// database is the common case, a database and a cache the next; the cap is what
// lets a creation dialog show everything a request is about to create.
const MaxDependencies = 3

// dependencyNamePattern is narrower than a parameter name. The name becomes the
// suffix of an app name, and that app's key is what environment references use,
// so it has to come through slugifying unchanged.
var dependencyNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]{0,15}$`)

// Dependency is another template of the same repository whose app is created
// alongside this template's own: a web app's database.
type Dependency struct {
	// Name is the role - db, cache - and the suffix of the created app's name.
	Name  string `yaml:"name"`
	Title string `yaml:"title"`
	// Template names a template in the same repository.
	Template string `yaml:"template"`
	// Version and Variant are empty for the dependency template's defaults.
	Version string `yaml:"version,omitempty"`
	Variant string `yaml:"variant,omitempty"`
	// Params are the values this template fixes for the dependency. A string may
	// use this template's own ${{ params.x }} placeholders.
	Params map[string]any `yaml:"params,omitempty"`
}

// DependencyAppName is the name of the app created for a dependency.
func DependencyAppName(appName, dependencyName string) string {
	return appName + "-" + dependencyName
}

func (t *Template) FindDependency(name string) *Dependency {
	for _, dep := range t.Dependencies {
		if dep != nil && dep.Name == name {
			return dep
		}
	}
	return nil
}

// AskedParams are the parameters of the dependency's template that the person
// creating the app fills in: those this dependency does not fix, with no default,
// not optional, and not secrets HivePaaS generates. For a database that is its
// data volume. An empty default counts as none, as it does when parameters are
// resolved.
func (d *Dependency) AskedParams(depTemplate *Template) []*Parameter {
	var asked []*Parameter
	for _, param := range depTemplate.Parameters {
		if param == nil {
			continue
		}
		if _, fixed := d.Params[param.Name]; fixed {
			continue
		}
		defaulted := param.Default != nil && param.Default != ""
		generated := param.Type == ParamTypeSecret && param.Generate != nil
		if defaulted || param.Optional || generated {
			continue
		}
		asked = append(asked, param)
	}
	return asked
}

// validateDependencies checks what a template says about its dependencies. What
// needs the templates they name is templaterepo's.
func validateDependencies(t *Template, p *problems) {
	if len(t.Dependencies) > MaxDependencies {
		p.add("dependencies: at most %d", MaxDependencies)
	}
	seen := map[string]bool{}
	for i, dep := range t.Dependencies {
		if dep == nil {
			p.add("dependencies[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("dependencies[%s]", dep.Name)
		if !dependencyNamePattern.MatchString(dep.Name) {
			p.add("dependencies[%d].name %q must be lowercase letters and digits, starting with a letter", i, dep.Name)
		}
		if seen[dep.Name] {
			p.add("%s is declared twice", prefix)
		}
		seen[dep.Name] = true
		if dep.Title == "" {
			p.add("%s.title is required", prefix)
		}
		switch {
		case !templateNamePattern.MatchString(dep.Template):
			p.add("%s.template %q is not a template name", prefix, dep.Template)
		case dep.Template == t.Metadata.Name:
			p.add("%s: a template cannot depend on itself", prefix)
		}
		if dep.Version != "" && !choiceNamePattern.MatchString(dep.Version) {
			p.add("%s.version %q is invalid", prefix, dep.Version)
		}
		if dep.Variant != "" && !choiceNamePattern.MatchString(dep.Variant) {
			p.add("%s.variant %q is invalid", prefix, dep.Variant)
		}
		for _, name := range slices.Sorted(maps.Keys(dep.Params)) {
			if !paramNamePattern.MatchString(name) {
				p.add("%s.params: %q is not a parameter name", prefix, name)
			}
		}
	}
}
```

- [ ] **Step 4: Add the field, the call and the index type**

In `template.go`, in `Template`, after `Parameters`:

```go
	Parameters []*Parameter `yaml:"parameters,omitempty"`
	// Dependencies are templates of the same repository whose apps are created
	// alongside this one.
	Dependencies []*Dependency `yaml:"dependencies,omitempty"`
```

In `validate.go`, in `Validate`, after `validateVersions(t.Versions, variants, &p)`:

```go
	validateDependencies(t, &p)
```

In `index.go`, in `IndexEntry`, after `License`:

```go
	// Dependencies let the store say what else a template creates without reading
	// its file.
	Dependencies []*IndexDependency `json:"dependencies,omitempty"`
```

and below `IndexVersion`:

```go
type IndexDependency struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	Template string `json:"template"`
}
```

Run `gofmt -w` on the three files.

- [ ] **Step 5: Run the package**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ -v -run 'Dependenc|AskedParams|Validate'`
Expected: PASS.

- [ ] **Step 6: Lint and commit**

Run: `golangci-lint run ./hivepaas_app/service/apptemplateservice/...` - expected `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templatemodel/
git commit -m "$(cat <<'EOF'
feat(apptemplate): a dependencies block in the template format

A template may name up to three templates of the same repository whose apps
are created alongside its own, each with a role name, an optional version
and variant, and the parameters it fixes. AskedParams says which of the
dependency's parameters a person still has to fill in.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: One list of what each kind shares

The renderer has to refuse `${{ deps.db.ref.HIVEPAAS_BUCKET }}` against a database, and only `envvarserviceimpl` knows what a kind publishes. The names move to `base`, and a test keeps the two from drifting apart.

**Files:**
- Modify: `hivepaas_app/base/env_var.go`
- Test: `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go`

**Interfaces:**
- Produces:
  - `var base.AppCommonSharedEnvVars []string` - shared by every app
  - `func base.AppKindSharedEnvVars(category base.AppCategory) []string` - shared on top of those by a kind

- [ ] **Step 1: Write the failing test**

Append to `app_system_env_test.go`:

```go
// What a kind publishes is read in two places: here, and by the template
// renderer deciding what ${{ deps.<name>.ref.VAR }} may name. This keeps them
// from drifting apart.
func TestKindEnvVars_SharedNamesMatchBase(t *testing.T) {
	kinds := map[base.AppCategory]*entity.AppKindSettings{
		base.AppCategoryDatabase: {Category: base.AppCategoryDatabase, Database: &entity.AppKindDatabase{}},
		base.AppCategoryCache:    {Category: base.AppCategoryCache, Cache: &entity.AppKindCache{}},
		base.AppCategoryStorage:  {Category: base.AppCategoryStorage, Storage: &entity.AppKindStorage{}},
		base.AppCategoryWebapp:   {Category: base.AppCategoryWebapp, Webapp: &entity.AppKindWebapp{}},
	}
	for _, category := range base.AllAppCategories {
		envs, err := kindEnvVars(kinds[category])
		assert.NoError(t, err)

		var shared []string
		for _, env := range envs {
			if env.IsShared {
				shared = append(shared, env.Key)
			}
		}
		assert.ElementsMatch(t, base.AppKindSharedEnvVars(category), shared, string(category))
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./hivepaas_app/service/envvarservice/envvarserviceimpl/ -run SharedNamesMatchBase -v`
Expected: FAIL to compile: `undefined: base.AppKindSharedEnvVars`.

- [ ] **Step 3: Add the lists to `base/env_var.go`**

Below the `const` block of the app system variables:

```go
// AppCommonSharedEnvVars are shared by every app, whatever its kind.
var AppCommonSharedEnvVars = []string{
	AppSystemEnvVarHost, AppSystemEnvVarPort, AppSystemEnvVarDomain,
	AppSystemEnvVarEnv, AppSystemEnvVarName, AppSystemEnvVarID,
}

// AppKindSharedEnvVars are what an app of a kind shares on top of the common
// variables - what another app's ${<app>.VAR} may name. envvarserviceimpl
// publishes exactly these, and a test there holds the two together.
func AppKindSharedEnvVars(category AppCategory) []string {
	switch category {
	case AppCategoryDatabase:
		return []string{AppSystemEnvVarUser, AppSystemEnvVarPassword, AppSystemEnvVarDatabaseName,
			AppSystemEnvVarSSLMode}
	case AppCategoryCache:
		return []string{AppSystemEnvVarPassword}
	case AppCategoryStorage:
		return []string{AppSystemEnvVarKeyID, AppSystemEnvVarSecret, AppSystemEnvVarBucket, AppSystemEnvVarRegion}
	case AppCategoryWebapp:
		return nil
	}
	return nil
}
```

- [ ] **Step 4: Run it**

Run: `go test ./hivepaas_app/service/envvarservice/... ./hivepaas_app/base/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/base/env_var.go hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go
git commit -m "$(cat <<'EOF'
feat(envvars): one list of the variables each app kind shares

The template renderer is about to refuse a reference to a variable a
dependency's kind does not share, and only the env var service knew what
that was. The names move to base, and a test holds the two together.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Rendering `deps.*` and a dependency's parameters

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templaterender/deps.go`
- Create: `hivepaas_app/service/apptemplateservice/templaterender/deps_test.go`
- Modify: `hivepaas_app/service/apptemplateservice/templaterender/render.go`

**Interfaces:**
- Consumes: `templatemodel.Dependency`, `(*Dependency).AskedParams`, `Template.Dependencies` (Task 2); `base.AppCommonSharedEnvVars`, `base.AppKindSharedEnvVars` (Task 3).
- Produces:
  - `type DepBinding struct { AppKey string; SharedVars []string }`
  - `Request.ResolvedParams map[string]*Value`, `Request.Deps map[string]*DepBinding`
  - `func DependencyParams(owner string, dep *templatemodel.Dependency, depTemplate *templatemodel.Template, ownerParams map[string]*Value, input map[string]any) (map[string]any, error)`
  - `func KindCategory(doc *specmodel.AppDoc) base.AppCategory`
  - `func SharedVarsOf(doc *specmodel.AppDoc) []string`

- [ ] **Step 1: Write the failing tests**

Create `deps_test.go`:

```go
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
	return map[string]*DepBinding{"db": {
		AppKey:     "blog_db",
		SharedVars: append(append([]string{}, base.AppCommonSharedEnvVars...), base.AppKindSharedEnvVars(base.AppCategoryDatabase)...),
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
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/ -run 'Dependenc|SharedVars|ResolvedParams' -v`
Expected: FAIL to compile: `undefined: DepBinding`.

- [ ] **Step 3: Write `deps.go`**

```go
package templaterender

import (
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// DepBinding ties a dependency a template declares to the app that serves it.
type DepBinding struct {
	// AppKey is the dependency app's key: what ${key.VAR} names.
	AppKey string
	// SharedVars are what that app shares. A reference to anything else is
	// refused, because it would resolve to nothing at deploy time.
	SharedVars []string
}

// resolveDep answers deps.<name>.key with the app key, and deps.<name>.ref.<VAR>
// with an environment reference to what that app shares.
func resolveDep(template, ref string, deps map[string]*DepBinding) (any, error) {
	rest := strings.TrimPrefix(ref, "deps.")
	name, field, found := strings.Cut(rest, ".")
	binding := deps[name]
	if !found || binding == nil {
		return nil, undefinedPlaceholder(template, ref)
	}
	if field == "key" {
		return binding.AppKey, nil
	}
	variable, isRef := strings.CutPrefix(field, "ref.")
	if !isRef || !slices.Contains(binding.SharedVars, variable) {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: placeholder %q names nothing dependency %q shares", template, ref, name)
	}
	return "${" + binding.AppKey + "." + variable + "}", nil
}

// DependencyParams builds a dependency's parameter input: the values the
// declaring template fixes, with its own parameter placeholders substituted, and
// what the person creating the app was asked. Anything else sent for the
// dependency is refused - the dialog and the template would disagree about what
// it needs.
func DependencyParams(
	owner string,
	dep *templatemodel.Dependency,
	depTemplate *templatemodel.Template,
	ownerParams map[string]*Value,
	input map[string]any,
) (map[string]any, error) {
	asked := map[string]bool{}
	for _, param := range dep.AskedParams(depTemplate) {
		asked[param.Name] = true
	}
	for _, name := range slices.Sorted(maps.Keys(input)) {
		if !asked[name] {
			return nil, paramInvalid(dep.Name+"."+name, "the dependency does not ask for it")
		}
	}

	fixed, err := Substitute(dep.Params, ownerParamResolver(owner, ownerParams))
	if err != nil {
		return nil, err
	}
	params := map[string]any{}
	maps.Copy(params, input)
	fixedParams, _ := fixed.(map[string]any)
	maps.Copy(params, fixedParams)
	return params, nil
}

// ownerParamResolver answers only params.* of the declaring template, and refuses
// a secret: a secret belongs to one app, and copying it into another's parameters
// would store it twice, possibly in the clear.
func ownerParamResolver(owner string, ownerParams map[string]*Value) Resolver {
	return func(ref string) (any, bool, error) {
		name, isParam := strings.CutPrefix(ref, "params.")
		value, found := ownerParams[name]
		if !isParam || !found {
			return nil, false, undefinedPlaceholder(owner, ref)
		}
		if value.Param.Type == templatemodel.ParamTypeSecret {
			return nil, false, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
				"%s: a dependency parameter cannot take the secret %q", owner, ref)
		}
		if value.Value == nil {
			return "", false, nil
		}
		return value.Value, false, nil
	}
}

// KindCategory is the kind a rendered document declares, empty when it declares
// none.
func KindCategory(doc *specmodel.AppDoc) base.AppCategory {
	kind, _ := doc.Settings[specmodel.SingletonBlockName(base.SettingTypeAppKind)].(map[string]any)
	category, _ := kind["category"].(string)
	return base.AppCategory(category)
}

// SharedVarsOf is everything an app built from doc shares with other apps.
func SharedVarsOf(doc *specmodel.AppDoc) []string {
	return append(slices.Clone(base.AppCommonSharedEnvVars), base.AppKindSharedEnvVars(KindCategory(doc))...)
}
```

- [ ] **Step 4: Wire it into `render.go`**

In `Request`, after `ImageOverride`:

```go
	// ResolvedParams, when set, are used instead of resolving Params. A caller that
	// renders a template's dependencies from its parameters resolves them first,
	// and this render must use the same generated secrets.
	ResolvedParams map[string]*Value
	// Deps binds every dependency the template declares to its app.
	Deps map[string]*DepBinding
```

In `Render`, replace

```go
	params, err := ResolveParams(tmpl.Parameters, req.Params)
	if err != nil {
		return nil, err
	}
```

with

```go
	params := req.ResolvedParams
	if params == nil {
		if params, err = ResolveParams(tmpl.Parameters, req.Params); err != nil {
			return nil, err
		}
	}
	for _, dep := range tmpl.Dependencies {
		if req.Deps[dep.Name] == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: dependency %q is not bound to an app", tmpl.Metadata.Name, dep.Name)
		}
	}
```

Change `resolver` to take the bindings, and answer `deps.` first:

```go
func resolver(
	template string,
	vars map[string]any,
	params map[string]*Value,
	deps map[string]*DepBinding,
	keepSecrets bool,
) Resolver {
	return func(ref string) (any, bool, error) {
		if strings.HasPrefix(ref, "deps.") {
			value, err := resolveDep(template, ref, deps)
			return value, false, err
		}
		if paramName, isParam := strings.CutPrefix(ref, "params."); isParam {
			// ... unchanged
```

and update its two call sites in `Render`:

```go
	applied, err := Substitute(tree, resolver(name, vars, params, req.Deps, false))
```

```go
	baseTree, err := Substitute(tree, resolver(name, vars, params, req.Deps, true))
```

- [ ] **Step 5: Run the package**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/`
Expected: `ok`.

- [ ] **Step 6: Lint and commit**

Run: `golangci-lint run ./hivepaas_app/service/apptemplateservice/...` - expected `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templaterender/
git commit -m "$(cat <<'EOF'
feat(apptemplate): render references to a template's dependencies

${{ deps.<name>.ref.VAR }} renders to ${<key>.VAR}, an ordinary environment
reference resolved at deploy time, and only for a variable the dependency's
kind shares. ${{ deps.<name>.key }} renders to the key alone.
DependencyParams builds a dependency's parameters from what the declaring
template fixes and what the person was asked, and refuses anything else -
including a secret of the declaring template.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Linting a template against the templates it names

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templaterepo/lint_dependencies.go`
- Create: `hivepaas_app/service/apptemplateservice/templaterepo/lint_dependencies_test.go`
- Modify: `hivepaas_app/service/apptemplateservice/templaterepo/lint.go` (`Lint`, `lintRenders`)
- Modify: `hivepaas_app/service/apptemplateservice/templaterepo/index.go` (`BuildIndex`)

**Interfaces:**
- Consumes: Task 2 (`Dependency`, `AskedParams`, `IndexDependency`), Task 4 (`DepBinding`, `DependencyParams`, `SharedVarsOf`, `Request.ResolvedParams`, `Request.Deps`).
- Produces: `func lintDependencies(repo *Repo, file *TemplateFile) []Problem`, `func lintBindings(repo *Repo, file *TemplateFile, owner map[string]*templaterender.Value) (map[string]*templaterender.DepBinding, []Problem)`; `lintRenders(repo *Repo, file *TemplateFile) []Problem`.

- [ ] **Step 1: Write the failing tests**

Create `lint_dependencies_test.go`:

```go
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

	var all []string
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
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterepo/ -run 'Dependenc|NestedDependency|ListsDependencies' -v`
Expected: FAIL - the valid case reports `dependency "db" is not bound to an app`, and the index has no dependencies.

- [ ] **Step 3: Write `lint_dependencies.go`**

```go
package templaterepo

import (
	"fmt"
	"maps"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// lintKeyPrefix makes the stand-in key a dependency is bound to while linting.
const lintKeyPrefix = "lint_"

// lintDependencies checks what a template's dependencies need from the rest of
// the repository. A problem here makes a bound render meaningless, so Lint skips
// the renders of a template that has one.
func lintDependencies(repo *Repo, file *TemplateFile) []Problem {
	var problems []Problem
	report := func(format string, args ...any) {
		problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(format, args...)})
	}
	for _, dep := range file.Template.Dependencies {
		prefix := "dependency " + dep.Name
		target := repo.FindTemplate(dep.Template)
		if target == nil {
			report("%s: template %q is not in this repository", prefix, dep.Template)
			continue
		}
		depTmpl := target.Template
		if len(depTmpl.Dependencies) > 0 {
			report("%s: template %q has dependencies of its own, and dependencies go one level deep",
				prefix, dep.Template)
		}
		if dep.Version != "" {
			version := depTmpl.FindVersion(dep.Version)
			switch {
			case version == nil:
				report("%s: template %q has no version %q", prefix, dep.Template, dep.Version)
			case version.Deprecated:
				report("%s: version %q of %q is deprecated", prefix, dep.Version, dep.Template)
			}
		}
		if dep.Variant != "" && depTmpl.FindVariant(dep.Variant) == nil {
			report("%s: template %q has no variant %q", prefix, dep.Template, dep.Variant)
		}
		for _, name := range slices.Sorted(maps.Keys(dep.Params)) {
			declared := slices.ContainsFunc(depTmpl.Parameters, func(param *templatemodel.Parameter) bool {
				return param != nil && param.Name == name
			})
			if !declared {
				report("%s: template %q declares no parameter %q", prefix, dep.Template, name)
			}
		}
	}
	return problems
}

// lintBindings renders each dependency the way creating an app would - its fixed
// parameters, a stand-in for each one a person is asked - and binds it under a
// stand-in key to what it shares.
func lintBindings(
	repo *Repo,
	file *TemplateFile,
	owner map[string]*templaterender.Value,
) (map[string]*templaterender.DepBinding, []Problem) {
	tmpl := file.Template
	bindings := map[string]*templaterender.DepBinding{}
	var problems []Problem
	for _, dep := range tmpl.Dependencies {
		depTmpl := repo.FindTemplate(dep.Template).Template
		asked := map[string]any{}
		for _, param := range dep.AskedParams(depTmpl) {
			asked[param.Name] = lintStandIn(param)
		}
		input, err := templaterender.DependencyParams(tmpl.Metadata.Name, dep, depTmpl, owner, asked)
		if err == nil {
			var result *templaterender.Result
			result, err = templaterender.Render(&templaterender.Request{
				Template: depTmpl, Version: dep.Version, Variant: dep.Variant, Params: input,
			})
			if err == nil {
				bindings[dep.Name] = &templaterender.DepBinding{
					AppKey: lintKeyPrefix + dep.Name, SharedVars: templaterender.SharedVarsOf(result.Doc),
				}
				continue
			}
		}
		problems = append(problems, Problem{Path: file.Path,
			Message: "dependency " + dep.Name + ": " + ErrorText(err)})
	}
	return bindings, problems
}

// lintStandIn is a value a person could have given for a parameter they are asked.
func lintStandIn(param *templatemodel.Parameter) any {
	switch param.Type {
	case templatemodel.ParamTypeVolume:
		return lintVolumeID
	case templatemodel.ParamTypeSecret:
		return lintSecret
	case templatemodel.ParamTypeInt, templatemodel.ParamTypeSize:
		if param.Min != nil {
			return param.Min
		}
		if param.Type == templatemodel.ParamTypeSize {
			return "1MB"
		}
		return 1
	case templatemodel.ParamTypeBool:
		return false
	case templatemodel.ParamTypeSelect:
		if len(param.Options) > 0 {
			return param.Options[0].Value
		}
		return ""
	case templatemodel.ParamTypeString:
		return "lint"
	}
	return "lint"
}
```

- [ ] **Step 4: Call it from `lint.go`**

In `Lint`, replace

```go
		problems = append(problems, lintRenders(file)...)
```

with

```go
		if depProblems := lintDependencies(repo, file); len(depProblems) > 0 {
			problems = append(problems, depProblems...)
			continue
		}
		problems = append(problems, lintRenders(repo, file)...)
```

Change `lintRenders` to bind dependencies before rendering. Its signature becomes `func lintRenders(repo *Repo, file *TemplateFile) []Problem`, and at the top of its body, after `tmpl := file.Template`:

```go
	request := func(version, variant string) *templaterender.Request {
		return &templaterender.Request{
			Template: tmpl, Version: version, Variant: variant,
			Params: lintParams(tmpl), AllowDeprecated: true,
		}
	}
	if len(tmpl.Dependencies) > 0 {
		owner, err := templaterender.ResolveParams(tmpl.Parameters, lintParams(tmpl))
		if err != nil {
			return []Problem{{Path: file.Path, Message: ErrorText(err)}}
		}
		deps, problems := lintBindings(repo, file, owner)
		if len(problems) > 0 {
			return problems
		}
		request = func(version, variant string) *templaterender.Request {
			return &templaterender.Request{
				Template: tmpl, Version: version, Variant: variant,
				ResolvedParams: owner, Deps: deps, AllowDeprecated: true,
			}
		}
	}
```

and inside its loop replace the `templaterender.Render(&templaterender.Request{...})` literal with:

```go
			_, err := templaterender.Render(request(version.Name, variant))
```

- [ ] **Step 5: Fill the index**

In `templaterepo/index.go`, in `BuildIndex`, after the `for _, variant := range tmpl.Variants` loop:

```go
		for _, dep := range tmpl.Dependencies {
			entry.Dependencies = append(entry.Dependencies,
				&templatemodel.IndexDependency{Name: dep.Name, Title: dep.Title, Template: dep.Template})
		}
```

- [ ] **Step 6: Run the package and the templates repository**

Run: `go test ./hivepaas_app/service/apptemplateservice/...`
Expected: `ok` for every package.

Run: `go run ./tools/apptemplate lint ../app-templates`
Expected: `OK: 21 template(s)` - nothing in the repository uses dependencies yet.

- [ ] **Step 7: Lint and commit**

Run: `golangci-lint run ./...` - expected `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templaterepo/
git commit -m "$(cat <<'EOF'
feat(apptemplate): lint a template against the templates it depends on

The linter refuses a dependency on a template the repository does not have,
on one with dependencies of its own, on a version or variant it does not
declare or has deprecated, and a parameter it does not declare. It then
renders each dependency with stand-ins for what a person is asked and
renders the template bound to them, which is what refuses a reference to a
variable the dependency does not share. The index lists dependencies.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: The service loads and renders dependencies

**Files:**
- Modify: `hivepaas_app/service/apptemplateservice/types.go`
- Modify: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go` (`Template`, `Render`)
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/templates/demoweb.yaml`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo/icons/demoweb.svg`
- Test: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service_test.go`

**Interfaces:**
- Consumes: Tasks 2 and 4.
- Produces:
  - `TemplateResp.Dependencies []*DependencyTemplate`; `type DependencyTemplate struct { Dependency *templatemodel.Dependency; Entry *templatemodel.IndexEntry; Template *templatemodel.Template }`
  - `RenderReq.AppName string`, `RenderReq.DependencyParams map[string]map[string]any`
  - `RenderResp.Dependencies []*RenderedDependency`; `type RenderedDependency struct { Name string; AppName string; Render *RenderResp }`

- [ ] **Step 1: Add the test template**

`testdata/repo/templates/demoweb.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: demoweb
  title: Demo Web
  tagline: A demo web app
  description: Demo web app.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demoweb.svg
  requires: {versionCode: v000001}
parameters:
  - {name: dataVolume, title: Site files, type: volume}
dependencies:
  - {name: db, title: Database, template: demo}
versions:
  - {name: "1", release: "1.0", default: true, image: "demoweb:1.0.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /srv: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    kind: {category: webapp, engine: demoweb, webapp: {}}
    envVars:
      data:
        - {k: DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
```

`testdata/repo/icons/demoweb.svg`:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"/>
```

- [ ] **Step 2: Write the failing tests**

Append to `service_test.go`:

```go
func TestServiceTemplateLoadsDependencies(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Template(context.Background(), "demoweb")

	assert.NoError(t, err)
	assert.Len(t, resp.Dependencies, 1)
	assert.Equal(t, "db", resp.Dependencies[0].Dependency.Name)
	assert.Equal(t, "demo", resp.Dependencies[0].Template.Metadata.Name)
	assert.Equal(t, "templates/demo.yaml", resp.Dependencies[0].Entry.File.Path)
}

func TestServiceRenderRendersDependenciesFirst(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name:             "demoweb",
		AppName:          "blog",
		Params:           map[string]any{"dataVolume": "vol-web"},
		DependencyParams: map[string]map[string]any{"db": {"dataVolume": "vol-db"}},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Dependencies, 1)
	db := resp.Dependencies[0]
	assert.Equal(t, "db", db.Name)
	assert.Equal(t, "blog-db", db.AppName)
	assert.Equal(t, "vol-db", db.Render.Result.Doc.Deployment.Storage.Mounts["/data"].Source)
	assert.Equal(t, localRevision, db.Render.Revision, "one revision for every app of the request")

	envVars := resp.Result.Doc.Settings["envVars"].(map[string]any)["data"].([]any)
	assert.Equal(t, "${blog_db.HIVEPAAS_PASSWORD}", envVars[0].(map[string]any)["v"])
}

func TestServiceRenderRefusesWhatADependencyDoesNotAsk(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	cases := map[string]*apptemplateservice.RenderReq{
		"no app name": {
			Name: "demoweb", Params: map[string]any{"dataVolume": "v"},
			DependencyParams: map[string]map[string]any{"db": {"dataVolume": "v"}},
		},
		"unknown dependency": {
			Name: "demoweb", AppName: "blog", Params: map[string]any{"dataVolume": "v"},
			DependencyParams: map[string]map[string]any{"cache": {"dataVolume": "v"}},
		},
		"dependency parameters for a template without dependencies": {
			Name: "demo", AppName: "blog", Params: map[string]any{"dataVolume": "v"},
			DependencyParams: map[string]map[string]any{"db": {"dataVolume": "v"}},
		},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Render(context.Background(), req)
			assert.Error(t, err)
		})
	}
}
```

- [ ] **Step 3: Run them and watch them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/ -run 'Dependenc|DoesNotAsk' -v`
Expected: FAIL to compile: `resp.Dependencies undefined`.

- [ ] **Step 4: Extend the types**

In `types.go`, replace `TemplateResp`, `RenderReq` and `RenderResp` with:

```go
type TemplateResp struct {
	Source   string
	Revision string
	Entry    *templatemodel.IndexEntry
	Template *templatemodel.Template
	// Dependencies are the templates this one depends on, loaded from the same
	// revision, in declaration order.
	Dependencies []*DependencyTemplate
}

type DependencyTemplate struct {
	Dependency *templatemodel.Dependency
	Entry      *templatemodel.IndexEntry
	Template   *templatemodel.Template
}
```

```go
type RenderReq struct {
	Name string
	// AppName is the name the app will have. A template with dependencies needs
	// it: each dependency's app is named after it, and referred to by key.
	AppName string
	Version string
	Variant string
	Params  map[string]any
	// DependencyParams are what the person was asked for each dependency, by the
	// dependency's name.
	DependencyParams map[string]map[string]any
	// ImageOverride is an image the user chose instead of the template's, empty to
	// use the template's own.
	ImageOverride string
}

type RenderResp struct {
	TemplateResp
	Result *templaterender.Result
	// Dependencies are rendered, and created, before the app that needs them.
	Dependencies []*RenderedDependency
}

type RenderedDependency struct {
	// Name is the dependency's role in the template that declares it.
	Name    string
	AppName string
	Render  *RenderResp
}
```

- [ ] **Step 5: Load dependencies in `Template`**

In `service.go`, replace `Template` with:

```go
func (s *service) Template(ctx context.Context, name string) (*apptemplateservice.TemplateResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	revision, err := src.Revision(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry, tmpl, err := loadTemplate(ctx, src, index, name)
	if err != nil {
		return nil, err
	}
	resp := &apptemplateservice.TemplateResp{Source: src.ID(), Revision: revision, Entry: entry, Template: tmpl}
	// The same index, so the same revision: a template and what it depends on are
	// never read from two different pins.
	for _, dep := range tmpl.Dependencies {
		depEntry, depTmpl, err := loadTemplate(ctx, src, index, dep.Template)
		if err != nil {
			return nil, err
		}
		resp.Dependencies = append(resp.Dependencies,
			&apptemplateservice.DependencyTemplate{Dependency: dep, Entry: depEntry, Template: depTmpl})
	}
	return resp, nil
}

func loadTemplate(
	ctx context.Context,
	src apptemplateservice.Source,
	index *templatemodel.Index,
	name string,
) (*templatemodel.IndexEntry, *templatemodel.Template, error) {
	entry := index.FindTemplate(name)
	if entry == nil {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateNotFound).WithParam("Name", name)
	}
	data, err := src.TemplateFile(ctx, entry)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	tmpl, err := templatemodel.DecodeTemplate(data)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return entry, tmpl, nil
}
```

- [ ] **Step 6: Render dependencies in `Render`**

Replace `Render` with:

```go
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
	if len(loaded.Dependencies) > 0 {
		return renderWithDependencies(loaded, req)
	}
	if len(req.DependencyParams) > 0 {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
			WithExtraDetail("%s has no dependencies to give parameters to", req.Name)
	}
	result, err := templaterender.Render(&templaterender.Request{
		Template:      loaded.Template,
		Version:       req.Version,
		Variant:       req.Variant,
		Params:        req.Params,
		ImageOverride: req.ImageOverride,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.RenderResp{TemplateResp: *loaded, Result: result}, nil
}

// renderWithDependencies renders every app of the request before any of them
// exists: the dependencies first, because the template refers to what they share
// and their parameters can refer to the template's.
func renderWithDependencies(
	loaded *apptemplateservice.TemplateResp,
	req *apptemplateservice.RenderReq,
) (*apptemplateservice.RenderResp, error) {
	tmpl := loaded.Template
	name := tmpl.Metadata.Name
	if req.AppName == "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: its dependencies are named after the app, and no app name was given", name)
	}
	for _, depName := range slices.Sorted(maps.Keys(req.DependencyParams)) {
		if tmpl.FindDependency(depName) == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
				WithExtraDetail("%s has no dependency %q", name, depName)
		}
	}
	owner, err := templaterender.ResolveParams(tmpl.Parameters, req.Params)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &apptemplateservice.RenderResp{TemplateResp: *loaded}
	bindings := make(map[string]*templaterender.DepBinding, len(loaded.Dependencies))
	keys := map[string]string{projecthelper.CalcAppKey(req.AppName): req.AppName}
	for _, dep := range loaded.Dependencies {
		rendered, binding, err := renderDependency(loaded, dep, req, owner, keys)
		if err != nil {
			return nil, err
		}
		bindings[dep.Dependency.Name] = binding
		resp.Dependencies = append(resp.Dependencies, rendered)
	}

	result, err := templaterender.Render(&templaterender.Request{
		Template:       tmpl,
		Version:        req.Version,
		Variant:        req.Variant,
		ResolvedParams: owner,
		Deps:           bindings,
		ImageOverride:  req.ImageOverride,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Result = result
	return resp, nil
}

// renderDependency renders one dependency and binds it to the key its app will
// have. keys holds the keys already taken by this request: slugifying truncates,
// so two long names can meet, and that is refused rather than left to the second
// provision to discover.
func renderDependency(
	loaded *apptemplateservice.TemplateResp,
	dep *apptemplateservice.DependencyTemplate,
	req *apptemplateservice.RenderReq,
	owner map[string]*templaterender.Value,
	keys map[string]string,
) (*apptemplateservice.RenderedDependency, *templaterender.DepBinding, error) {
	name := loaded.Template.Metadata.Name
	depTmpl := dep.Template
	if !templatemodel.IsCompatible(depTmpl.Metadata.Requires, base.CurrentVersion) {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).WithParam("Name", depTmpl.Metadata.Name)
	}
	if len(depTmpl.Dependencies) > 0 {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"%s: dependency %q has dependencies of its own", name, dep.Dependency.Name)
	}

	appName := templatemodel.DependencyAppName(req.AppName, dep.Dependency.Name)
	appKey := projecthelper.CalcAppKey(appName)
	if other, taken := keys[appKey]; taken {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"%s: the apps %q and %q would have the same key; choose a shorter name", name, other, appName)
	}
	keys[appKey] = appName

	input, err := templaterender.DependencyParams(name, dep.Dependency, depTmpl, owner,
		req.DependencyParams[dep.Dependency.Name])
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	result, err := templaterender.Render(&templaterender.Request{
		Template: depTmpl, Version: dep.Dependency.Version, Variant: dep.Dependency.Variant, Params: input,
	})
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	rendered := &apptemplateservice.RenderedDependency{
		Name:    dep.Dependency.Name,
		AppName: appName,
		Render: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{
				Source: loaded.Source, Revision: loaded.Revision, Entry: dep.Entry, Template: depTmpl,
			},
			Result: result,
		},
	}
	binding := &templaterender.DepBinding{AppKey: appKey, SharedVars: templaterender.SharedVarsOf(result.Doc)}
	return rendered, binding, nil
}
```

Add the imports `maps`, `slices` and `github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper`.

- [ ] **Step 7: Run the package**

Run: `go test ./hivepaas_app/service/apptemplateservice/...`
Expected: `ok` for every package. If `projecthelper` creates an import cycle, `go build` names it; `projecthelper` depends on `base` and `slugify` only, so none is expected.

- [ ] **Step 8: Lint and commit**

Run: `golangci-lint run ./...` - expected `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/
git commit -m "$(cat <<'EOF'
feat(apptemplate): load and render a template's dependencies

Template loads the templates a template depends on from the same index, so
never from two pins. Render renders every app of a request before any
exists: each dependency with the parameters the template fixes and the
person was asked, bound to the key its app will have, then the template
itself with the same resolved parameters. Two apps of one request whose
names slugify to the same key are refused.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Ids chosen up front, and the links on the binding

**Files:**
- Modify: `hivepaas_app/service/appprovisionservice/types.go` (`ProvisionAppReq`)
- Modify: `hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/provision.go` (`newApp`)
- Modify: `hivepaas_app/entity/setting_app_template.go` (`AppTemplateSettings`)

**Interfaces:**
- Produces:
  - `ProvisionAppReq.AppID string`
  - `AppTemplateSettings.Dependencies []AppTemplateDependency`, `AppTemplateSettings.CreatedForAppID string`
  - `type AppTemplateDependency struct { Name, AppID, Template string }` (json `name`, `appId`, `template`)

- [ ] **Step 1: `ProvisionAppReq.AppID`**

In `appprovisionservice/types.go`, in `ProvisionAppReq`, before `Name`:

```go
	// AppID is the id the app is created with, generated when empty. The apps a
	// template creates together name each other by id, so their ids are chosen
	// before the first of them exists.
	AppID string
```

In `provision.go`, in `newApp`, replace `ID: gofn.Must(ulid.NewStringULID()),` with `ID: id,` and add before the literal:

```go
	id := req.AppID
	if id == "" {
		id = gofn.Must(ulid.NewStringULID())
	}
```

- [ ] **Step 2: The links on the binding**

In `entity/setting_app_template.go`, in `AppTemplateSettings`, before `Params`:

```go
	// Dependencies are the apps created alongside this one because its template
	// named them, in the order they were created.
	Dependencies []AppTemplateDependency `json:"dependencies,omitempty"`
	// CreatedForAppID is the app this one was created to serve, empty when it was
	// created on its own. Deleting that app leaves this one where it is.
	CreatedForAppID string `json:"createdForAppId,omitempty"`
```

and below `AppTemplateParam`:

```go
type AppTemplateDependency struct {
	// Name is the role the template gave it: db, cache.
	Name     string `json:"name"`
	AppID    string `json:"appId"`
	Template string `json:"template"`
}
```

- [ ] **Step 3: Build and test**

Run: `go build ./... && go test ./hivepaas_app/entity/... ./hivepaas_app/service/appprovisionservice/...`
Expected: build succeeds, tests `ok`. Both changes are additive; the use case in Task 8 is what exercises them.

- [ ] **Step 4: Commit**

```bash
git add hivepaas_app/service/appprovisionservice/ hivepaas_app/entity/setting_app_template.go
git commit -m "$(cat <<'EOF'
feat(apptemplate): choose an app's id before it exists, and link bindings

ProvisionAppReq takes an optional id, so the apps one template creates can
name each other before the first is provisioned. The app-template binding
records the dependencies an app was created with, and on a dependency, the
app it was created for.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: Creating every app of a request in one transaction

**Files:**
- Modify: `hivepaas_app/usecase/apptemplateuc/app_create_from_template.go`
- Modify: `hivepaas_app/usecase/apptemplateuc/audit.go`
- Test: `hivepaas_app/usecase/apptemplateuc/app_create_from_template_test.go`

**Interfaces:**
- Consumes: Task 6 (`RenderReq.AppName`, `RenderReq.DependencyParams`, `RenderResp.Dependencies`), Task 7 (`ProvisionAppReq.AppID`, `AppTemplateDependency`, `CreatedForAppID`), Task 9's request field `CreateAppFromTemplateReq.DependencyParams` (added in Step 1 below, since this task needs it first).
- Produces:
  - `type appToProvision struct { id, name, role string; rendered *apptemplateservice.RenderResp; links appTemplateLinks }`
  - `type appTemplateLinks struct { dependencies []entity.AppTemplateDependency; createdForAppID string }`
  - `func planApps(req *apptemplatedto.CreateAppFromTemplateReq, rendered *apptemplateservice.RenderResp) []*appToProvision`
  - `func (uc *UC) provisionAll(ctx, db, auth, req, apps []*appToProvision) ([]*createdFromTemplate, error)`
  - `func (uc *UC) removeServices(ctx context.Context, created []*createdFromTemplate) error`
  - `provisionFromTemplate(ctx, db, auth, req, target *appToProvision)`; `createdFromTemplate.role string`

- [ ] **Step 1: Add the request field**

In `apptemplatedto/app_create_from_template.go`, in `CreateAppFromTemplateReq`, after `Params`:

```go
	// DependencyParams are the values a person gave for each dependency the
	// template declares, by the dependency's name: in practice its data volume.
	DependencyParams map[string]map[string]any `json:"dependencyParams"`
```

- [ ] **Step 2: Write the failing tests**

In `app_create_from_template_test.go`:

1. Replace `fakeProvisionService.ProvisionApp` so each app it provisions is distinct:

```go
func (f *fakeProvisionService) ProvisionApp(
	ctx context.Context, db database.IDB, req *appprovisionservice.ProvisionAppReq,
) (*appprovisionservice.ProvisionAppResp, error) {
	f.called = true
	f.names = append(f.names, req.Name)
	if f.failName != "" && req.Name == f.failName {
		return nil, errTestProvision
	}
	id := req.AppID
	if id == "" {
		id = "app-1"
	}
	app := &entity.App{
		ID: id, Name: req.Name, Key: req.Name, ServiceID: "svc-" + req.Name,
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
```

and give the fake the fields it now uses:

```go
type fakeProvisionService struct {
	appprovisionservice.Service
	called   bool
	names    []string
	failName string
}
```

with `var errTestProvision = errors.New("provisioning failed")` beside `errTestRouting`.

2. `provisionFromTemplate` now takes the app to provision instead of the rendered response. Add a helper that builds the one the existing tests mean, and use it at all three call sites - the `provision` helper and the two routing tests that call `provisionFromTemplate` directly with `fakes.templates.resp`:

```go
// singleApp is the request's own app, with the id the fake provisions under.
func singleApp(fakes *createFakes) *appToProvision {
	return &appToProvision{id: "app-1", name: testCreateReq().Name, rendered: fakes.templates.resp}
}

func provision(t *testing.T, uc *UC, fakes *createFakes) *createdFromTemplate {
	t.Helper()
	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		singleApp(fakes))
	assert.NoError(t, err)
	return created
}
```

In the two routing tests, replace the last argument `fakes.templates.resp` of `uc.provisionFromTemplate(...)` with `singleApp(fakes)`.

Add a helper for reading an error's explanation - `Error()` of an hperrors error is only its code:

```go
func errDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}
```

3. Add a rendered request with a dependency, and the tests:

```go
// renderedWithDependency is what the service returns for a web app whose
// database is the test template.
func renderedWithDependency(t *testing.T, fakes *createFakes) *apptemplateservice.RenderResp {
	t.Helper()
	database := fakes.templates.resp
	web := *database
	web.Template = &templatemodel.Template{Metadata: templatemodel.Metadata{Name: "blog", Title: "Blog"}}
	web.Dependencies = []*apptemplateservice.RenderedDependency{
		{Name: "db", AppName: "blog-db", Render: database},
	}
	return &web
}

func TestPlanAppsPutsDependenciesFirstAndLinksThem(t *testing.T) {
	_, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "blog"

	apps := planApps(req, renderedWithDependency(t, fakes))

	assert.Len(t, apps, 2)
	db, blog := apps[0], apps[1]
	assert.Equal(t, "blog-db", db.name)
	assert.Equal(t, "db", db.role)
	assert.Equal(t, blog.id, db.links.createdForAppID)
	assert.Equal(t, "blog", blog.name)
	assert.Equal(t, []entity.AppTemplateDependency{{Name: "db", AppID: db.id, Template: "pg"}},
		blog.links.dependencies)
}

func TestProvisionAllRecordsBothDirections(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "blog"

	created, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, renderedWithDependency(t, fakes)))

	assert.NoError(t, err)
	assert.Equal(t, []string{"blog-db", "blog"}, fakes.provision.names, "the database first")
	binding := func(app *entity.App) *entity.AppTemplateSettings {
		setting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppTemplate)
		stored := &entity.Setting{Type: base.SettingTypeAppTemplate, Data: setting.Data}
		return stored.MustAsAppTemplateSettings()
	}
	db, blog := created[0].app, created[1].app
	assert.Equal(t, blog.ID, binding(db).CreatedForAppID)
	assert.Equal(t, db.ID, binding(blog).Dependencies[0].AppID)
	assert.Len(t, fakes.audit.entries, 2)
	assert.Contains(t, fakes.audit.entries[1].Detail, db.ID, "the app's creation names its database")
}

func TestProvisionAllNamesTheDependencyThatFailed(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.provision.failName = "blog-db"
	req := testCreateReq()
	req.Name = "blog"

	created, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, renderedWithDependency(t, fakes)))

	assert.ErrorIs(t, err, errTestProvision)
	assert.Contains(t, errDetail(t, err), "blog-db")
	assert.Empty(t, created)
}

type fakeClusterService struct {
	clusterservice.Service
	removed []string
	failID  string
}

func (f *fakeClusterService) ServiceRemove(_ context.Context, serviceID string, _ int, _ time.Duration) error {
	f.removed = append(f.removed, serviceID)
	if serviceID == f.failID {
		return errTestProvision
	}
	return nil
}

func TestRemoveServicesNewestFirstAndNamesWhatStays(t *testing.T) {
	cluster := &fakeClusterService{failID: "svc-blog-db"}
	uc := &UC{clusterService: cluster}
	created := []*createdFromTemplate{
		{app: &entity.App{Name: "blog-db", ServiceID: "svc-blog-db"}},
		{app: &entity.App{Name: "blog", ServiceID: "svc-blog"}},
	}

	err := uc.removeServices(context.Background(), created)

	assert.Equal(t, []string{"svc-blog", "svc-blog-db"}, cluster.removed)
	assert.ErrorIs(t, err, errTestProvision)
	assert.Contains(t, errDetail(t, err), "blog-db", "an orphan nobody is told about is worse than an orphan")
}
```

Add the imports `github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper` and `github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice` to the test file.

- [ ] **Step 3: Run them and watch them fail**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/ -run 'PlanApps|ProvisionAll|RemoveServices|ProvisionFromTemplate' -v`
Expected: FAIL to compile: `undefined: appToProvision`.

- [ ] **Step 4: Implement the planning and provisioning**

In `app_create_from_template.go`:

Add beside `createdFromTemplate`, and give that struct a `role` field:

```go
type createdFromTemplate struct {
	// role is the dependency's name, empty for the app the request asked for.
	role           string
	app            *entity.App
	deployment     *entity.Deployment
	deploymentTask *entity.Task
}

// appToProvision is one app of a creation request: the one it asked for, or a
// dependency created for it.
type appToProvision struct {
	id       string
	name     string
	role     string
	rendered *apptemplateservice.RenderResp
	links    appTemplateLinks
}

// appTemplateLinks are what an app's binding records about the other apps of
// the same request.
type appTemplateLinks struct {
	dependencies    []entity.AppTemplateDependency
	createdForAppID string
}

// planApps lists the apps a request creates, dependencies first. Their ids are
// chosen here, so that each binding can name the others before any exists.
func planApps(
	req *apptemplatedto.CreateAppFromTemplateReq,
	rendered *apptemplateservice.RenderResp,
) []*appToProvision {
	main := &appToProvision{id: gofn.Must(ulid.NewStringULID()), name: req.Name, rendered: rendered}
	apps := make([]*appToProvision, 0, len(rendered.Dependencies)+1)
	for _, dep := range rendered.Dependencies {
		depApp := &appToProvision{
			id:       gofn.Must(ulid.NewStringULID()),
			name:     dep.AppName,
			role:     dep.Name,
			rendered: dep.Render,
			links:    appTemplateLinks{createdForAppID: main.id},
		}
		main.links.dependencies = append(main.links.dependencies, entity.AppTemplateDependency{
			Name: dep.Name, AppID: depApp.id, Template: dep.Render.Template.Metadata.Name,
		})
		apps = append(apps, depApp)
	}
	return append(apps, main)
}

// provisionAll provisions the apps of a request in order. It returns what it
// created even when it fails part way, so the caller can remove their services.
func (uc *UC) provisionAll(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) ([]*createdFromTemplate, error) {
	created := make([]*createdFromTemplate, 0, len(apps))
	for _, target := range apps {
		one, err := uc.provisionFromTemplate(ctx, db, auth, req, target)
		if one != nil {
			created = append(created, one)
		}
		if err != nil {
			if target.role != "" {
				return created, hperrors.Wrap(err).WithExtraDetail(
					"while creating %s, which the template creates for %s", target.name, req.Name)
			}
			return created, hperrors.Wrap(err)
		}
	}
	return created, nil
}

// removeServices removes the services of apps whose transaction did not commit,
// newest first. Their records were rolled back, so a service that cannot be
// removed is running with nothing recorded, and the error names it.
func (uc *UC) removeServices(ctx context.Context, created []*createdFromTemplate) error {
	var errs []error
	for i := len(created) - 1; i >= 0; i-- {
		app := created[i].app
		if app == nil || app.ServiceID == "" {
			continue
		}
		err := uc.clusterService.ServiceRemove(ctx, app.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
		if err != nil {
			errs = append(errs, hperrors.Wrap(err).WithExtraDetail(
				"%s was created and could not be removed: remove service %s by hand", app.Name, app.ServiceID))
		}
	}
	return errors.Join(errs...)
}
```

Change `provisionFromTemplate` to take a target. Its signature and the parts that change:

```go
func (uc *UC) provisionFromTemplate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	target *appToProvision,
) (*createdFromTemplate, error) {
	timeNow := timeutil.NowUTC()
	rendered := target.rendered
	provisioned, err := uc.appProvisionService.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
		ProjectID:    req.ProjectID,
		ProjectEnvID: req.ProjectEnvID,
		AppID:        target.id,
		Name:         target.name,
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
			binding, bindErr := newAppTemplateSetting(app, rendered, target.links, timeNow)
			if bindErr != nil {
				return nil, hperrors.Wrap(bindErr)
			}
			return append(built.Settings, binding), nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	created := &createdFromTemplate{role: target.role, app: provisioned.App}
	app := created.app

	if err = uc.applyEnvVars(ctx, db, app); err != nil {
		return created, hperrors.Wrap(err).WithExtraDetail("while applying environment variables")
	}
	if err = uc.applyRouting(ctx, db, app); err != nil {
		return created, hperrors.Wrap(err).WithExtraDetail("while applying routing settings")
	}
	if err = uc.createFirstDeployment(ctx, db, auth, created); err != nil {
		return created, hperrors.Wrap(err)
	}
	return created, uc.recordCreateFromTemplate(ctx, db, auth, app, rendered, target.links)
}
```

The environment of every app in the scope is rebuilt each time an app is provisioned, so when the app's own turn comes its references to the database created a moment earlier in the same transaction resolve.

In `newAppTemplateSetting`, add the `links appTemplateLinks` parameter after `rendered`, and after `data.ImageOverride = result.ImageOverride`:

```go
	data.Dependencies = links.dependencies
	data.CreatedForAppID = links.createdForAppID
```

- [ ] **Step 5: Use it in `CreateAppFromTemplate`**

Replace the body after the `Render` call's error check with:

```go
	apps := planApps(req, rendered)
	for _, target := range apps {
		if doc := target.rendered.Result.Doc; doc.Deployment == nil || doc.Deployment.Source == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: the template deploys no image", target.rendered.Template.Metadata.Name)
		}
	}

	var created []*createdFromTemplate
	committed := false
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		// The swarm services are created inside the transaction but are not part
		// of it: a rolled-back request would leave them running with nothing
		// recorded.
		if err != nil && !committed {
			if removeErr := uc.removeServices(context.WithoutCancel(ctx), created); removeErr != nil {
				err = errors.Join(err, removeErr)
			}
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		created, txErr = uc.provisionAll(ctx, db, auth, req, apps)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed. The dependencies' deployments go first.
	for _, one := range created {
		if err = uc.taskQueue.ScheduleTask(ctx, one.deploymentTask); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return transformCreated(created), nil
}

// transformCreated describes what a request created: the app it asked for is
// the last one, and the others are its dependencies.
func transformCreated(created []*createdFromTemplate) *apptemplatedto.CreateAppFromTemplateResp {
	data := &apptemplatedto.CreateAppFromTemplateDataResp{
		Dependencies: make([]*apptemplatedto.CreatedDependencyResp, 0, len(created)),
	}
	for _, one := range created {
		app := &basedto.ObjectIDResp{ID: one.app.ID}
		deployment := &basedto.ObjectIDResp{ID: one.deployment.ID}
		if one.role == "" {
			data.App, data.Deployment = app, deployment
			continue
		}
		data.Dependencies = append(data.Dependencies,
			&apptemplatedto.CreatedDependencyResp{Name: one.role, App: app, Deployment: deployment})
	}
	return &apptemplatedto.CreateAppFromTemplateResp{Data: data}
}
```

and pass the new fields to `Render`:

```go
	rendered, err := uc.appTemplateService.Render(ctx, &apptemplateservice.RenderReq{
		Name:             req.Template,
		AppName:          req.Name,
		Version:          req.Version,
		Variant:          req.Variant,
		Params:           req.Params,
		DependencyParams: req.DependencyParams,
		ImageOverride:    req.ImageOverride,
	})
```

In `apptemplatedto/app_create_from_template.go`, add to `CreateAppFromTemplateDataResp`:

```go
	// Dependencies are the apps created alongside, each with its first deployment.
	Dependencies []*CreatedDependencyResp `json:"dependencies"`
```

and the type:

```go
type CreatedDependencyResp struct {
	// Name is the dependency's role in the template: db, cache.
	Name       string                `json:"name"`
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}
```

- [ ] **Step 6: Record the links**

In `audit.go`, add the `links appTemplateLinks` parameter to `recordCreateFromTemplate` after `rendered`, and before `err := auditservice.RecordAllowed(...)`:

```go
	if len(links.dependencies) > 0 {
		deps := make([]map[string]string, 0, len(links.dependencies))
		for _, dep := range links.dependencies {
			deps = append(deps, map[string]string{"name": dep.Name, "appId": dep.AppID, "template": dep.Template})
		}
		detail.Set("dependencies", deps)
	}
	detail.Set("createdFor", links.createdForAppID)
```

(`Set` ignores an empty value, so an app created on its own records no `createdFor`.)

- [ ] **Step 7: Run the package**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/...`
Expected: `ok`. The existing tests keep passing with the `provision` helper from Step 2.

- [ ] **Step 8: Lint and commit**

Run: `golangci-lint run ./...` - expected `0 issues.`

```bash
git add hivepaas_app/usecase/apptemplateuc/
git commit -m "$(cat <<'EOF'
feat(apptemplate): create a template's dependencies with the app

A request creates its dependencies first and its own app last, in one
transaction, with every id chosen up front so each binding and each audit
entry names the others. When anything fails, the services created so far
are removed newest first; one that cannot be removed is named in the error,
since its record was rolled back and nothing else would say it is running.
Deployments are scheduled after the commit, the dependencies' first.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 9: The wire format

**Files:**
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/app_create_from_template.go` (`Validate`)
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/template_list.go` (`AppTemplateSummaryResp`, `transformSummary`)
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/template_get.go` (`AppTemplateResp`, `TransformAppTemplate`)
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/template_binding_get.go`
- Test: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/transform_test.go`
- Regenerate: `docs/openapi/swagger.json`

**Interfaces:**
- Consumes: Task 2 (`IndexEntry.Dependencies`, `AskedParams`), Task 6 (`TemplateResp.Dependencies`), Task 7 (binding fields).
- Produces (JSON):
  - list: `dependencies: [{name, title, template}]`
  - get: `dependencies: [{name, title, template, templateTitle, version, variant, parameters: [...]}]` - `parameters` are only the asked ones
  - binding: `dependencies: [{name, appId, template}]`, `createdForAppId`

- [ ] **Step 1: Write the failing tests**

Append to `transform_test.go`:

```go
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
		Entry:    testEntry("v000001"),
		Template: &templatemodel.Template{Metadata: templatemodel.Metadata{Name: "wordpress"}, Dependencies: []*templatemodel.Dependency{dep}},
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
```

Add the import `github.com/hivepaas/hivepaas/hivepaas_app/entity`.

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/apptemplatedto/ -v`
Expected: FAIL to compile: `undefined: AppTemplateDependencySummaryResp`.

- [ ] **Step 3: The list summary**

In `template_list.go`, in `AppTemplateSummaryResp` after `License`:

```go
	// Dependencies are the apps creating this template also creates.
	Dependencies []*AppTemplateDependencySummaryResp `json:"dependencies"`
```

the type:

```go
type AppTemplateDependencySummaryResp struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	Template string `json:"template"`
}
```

and in `transformSummary`, in the literal `Dependencies: make([]*AppTemplateDependencySummaryResp, 0, len(entry.Dependencies)),`, then after the versions loop:

```go
	for _, dep := range entry.Dependencies {
		summary.Dependencies = append(summary.Dependencies,
			&AppTemplateDependencySummaryResp{Name: dep.Name, Title: dep.Title, Template: dep.Template})
	}
```

- [ ] **Step 4: The template**

In `template_get.go`, in `AppTemplateResp` after `Parameters`:

```go
	// Dependencies are the apps creating this template also creates, each with the
	// parameters a person has to fill in for it.
	Dependencies []*AppTemplateDependencyResp `json:"dependencies"`
```

the type:

```go
type AppTemplateDependencyResp struct {
	Name          string `json:"name"`
	Title         string `json:"title"`
	Template      string `json:"template"`
	TemplateTitle string `json:"templateTitle"`
	// Version and Variant are empty for the dependency template's defaults.
	Version string `json:"version"`
	Variant string `json:"variant"`
	// Parameters are only those the person is asked: the template fixes the rest,
	// or they have defaults, or HivePaaS generates them.
	Parameters []*AppTemplateParamResp `json:"parameters"`
}
```

and at the end of `TransformAppTemplate`, before `return resp`:

```go
	resp.Dependencies = make([]*AppTemplateDependencyResp, 0, len(tmpl.Dependencies))
	for _, dep := range tmpl.Dependencies {
		asked := dep.Dependency.AskedParams(dep.Template)
		depResp := &AppTemplateDependencyResp{
			Name:          dep.Dependency.Name,
			Title:         dep.Dependency.Title,
			Template:      dep.Dependency.Template,
			TemplateTitle: dep.Template.Metadata.Title,
			Version:       dep.Dependency.Version,
			Variant:       dep.Dependency.Variant,
			Parameters:    make([]*AppTemplateParamResp, 0, len(asked)),
		}
		for _, param := range asked {
			depResp.Parameters = append(depResp.Parameters, transformParam(param))
		}
		resp.Dependencies = append(resp.Dependencies, depResp)
	}
```

- [ ] **Step 5: The binding and the request validation**

In `template_binding_get.go`, in `AppTemplateBindingResp` after `ImageOverride`:

```go
	// Dependencies are the apps created alongside this one.
	Dependencies []*AppTemplateBindingDependencyResp `json:"dependencies"`
	// CreatedForAppID is the app this one was created to serve, empty when it was
	// created on its own.
	CreatedForAppID string `json:"createdForAppId"`
```

the type:

```go
type AppTemplateBindingDependencyResp struct {
	Name     string `json:"name"`
	AppID    string `json:"appId"`
	Template string `json:"template"`
}
```

and in `TransformAppTemplateBinding`, replace the `return &AppTemplateBindingResp{...}` with an assignment `resp := &AppTemplateBindingResp{...}` carrying the same fields plus `CreatedForAppID: settings.CreatedForAppID` and `Dependencies: make([]*AppTemplateBindingDependencyResp, 0, len(settings.Dependencies))`, then:

```go
	for _, dep := range settings.Dependencies {
		resp.Dependencies = append(resp.Dependencies,
			&AppTemplateBindingDependencyResp{Name: dep.Name, AppID: dep.AppID, Template: dep.Template})
	}
	return resp
```

In `app_create_from_template.go`, in `Validate`, raise the capacity hint from 7 to 8 and add before the `return`:

```go
	validators = append(validators,
		basedto.ValidateCond(len(req.DependencyParams) <= templatemodel.MaxDependencies, "dependencyParams")...)
```

(`basedto.ValidateCond(mustCond bool, field string)` is the helper `appsettingsdto` uses; add the `templatemodel` import.)

- [ ] **Step 6: Run, regenerate, lint**

Run: `go test ./hivepaas_app/usecase/apptemplateuc/...` - expected `ok`.
Run: `make gen-swag` - expected to rewrite `docs/openapi/swagger.json` with the new fields.
Run: `golangci-lint run ./...` - expected `0 issues.`
Run: `go test ./...` - expected no failures.

- [ ] **Step 7: Commit**

```bash
git add hivepaas_app/usecase/apptemplateuc/apptemplatedto/ docs/openapi/swagger.json
git commit -m "$(cat <<'EOF'
feat(apptemplate): dependencies on the wire

The list says what else a template creates. The template response lists
each dependency with only the parameters a person has to fill in for it.
Creating takes dependencyParams and answers with the dependency apps and
their deployments, and an app's template binding names its dependencies or
the app it was created for.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 10: Bring the spec in line with what was built

Two places in the spec say something the implementation does better, and a spec that disagrees with the code misleads whoever reads it next.

**Files:**
- Modify: `docs/superpowers/specs/2026-09-17-app-template-dependencies-design.md` (§4, §6, §8)

- [ ] **Step 1: §4 - everything is rendered before anything is created**

Replace the numbered list in §4 with:

```markdown
1. The template's parameters are resolved, so the dependencies' `params` can refer to them.
2. Each dependency, in declaration order, is loaded from the same index - so from the same
   revision - and rendered from what the template fixes plus `dependencyParams`, bound to
   the key its app will have: the slug of `<app name>-<dependency name>`.
3. The template is rendered with `deps` bound to those keys and to what each dependency's
   kind shares.
4. Only then, in one transaction, the apps are provisioned - dependencies first - each exactly
   as a direct creation would be: settings, environment, routing, first deployment. Their
   ids are chosen before the first one exists, so each binding can name the others.
```

and replace the paragraph beginning "The key of a created dependency is the slug of its name" with:

```markdown
A request whose apps would share a key - slugifying truncates long names - is refused while
rendering, before anything is created. A name already taken by an app in the project is
refused when that app is provisioned, and the apps provisioned before it in the same request
are removed with the transaction.
```

- [ ] **Step 2: §6 - the index**

At the end of §6's first paragraph, add:

```markdown
`index.json` is read by every installation of a channel whatever its version, so it is
decoded tolerantly: an older HivePaaS ignores `dependencies` rather than refusing the whole
index. Template files stay strict, and `requires.versionCode` keeps an older HivePaaS from
provisioning a template it cannot read.
```

- [ ] **Step 3: §8 - secrets stay with their app**

Add a paragraph at the end of §8:

```markdown
A dependency's `params` may take the declaring template's parameters, but never one of its
secrets: a secret belongs to one app, and copying it into another's parameters would store it
twice.
```

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/specs/2026-09-17-app-template-dependencies-design.md
git commit -m "$(cat <<'EOF'
docs: align the dependencies spec with the implementation

Every app of a request is rendered before any is created, a key collision
inside the request is refused while rendering, the index is read
tolerantly, and a dependency never receives the declaring template's
secrets.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 11: WordPress, the first template with a dependency

**Files (in `../app-templates`):**
- Create: `templates/wordpress.yaml`
- Create: `icons/wordpress.svg`
- Modify: `tags.yaml`, `README.md`, `index.json`

- [ ] **Step 1: Branch and confirm the image**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/app-templates
git checkout feat/database-templates
git checkout -b feat/webapp-templates
docker manifest inspect wordpress:7.1.0-php8.2-apache | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(sorted({m['platform']['architecture'] for m in d.get('manifests',[])}))"
```

Expected: a list containing `amd64` and `arm64`. If the tag no longer exists, pick the newest `X.Y.Z-phpA.B-apache` tag from `https://hub.docker.com/v2/repositories/library/wordpress/tags/?page_size=100&ordering=last_updated` and use it in the next step, `release` included.

- [ ] **Step 2: The icon and the tag**

```bash
cp ../hivepaas/assets/icons/wordpress.svg icons/wordpress.svg
python3 - <<'PY'
p = 'tags.yaml'
s = open(p).read()
if 'id: cms,' not in s:
    s = s.rstrip('\n') + '\n  - {id: cms, title: CMS}\n'
open(p, 'w').write(s)
PY
```

- [ ] **Step 3: Write `templates/wordpress.yaml`**

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate

metadata:
  name: wordpress
  title: WordPress
  tagline: The publishing platform behind a large share of the web
  description: |
    WordPress runs blogs, company sites and shops, extended by tens of thousands of themes and
    plugins.

    **This template creates two apps**

    - **WordPress**, on a volume for its code, themes, plugins and uploads
    - **a MySQL database** for it, named after the app with `-db` added - `blog-db` for an app
      called `blog` - with generated credentials

    The two are wired together through environment references, so the database password is
    never copied into WordPress's settings. Open the app's address after the first deployment
    to run the installer.

    **Deleting**

    Deleting the WordPress app leaves its database where it is: it holds your content, and a
    click aimed at one app should not destroy another's data. Delete the database separately
    once you no longer need it.

    **Starting up**

    WordPress may start before its database accepts connections. It answers with a database
    error until then, fails its health check, and is restarted until the database is ready -
    usually within a minute of creating both.
  categories: [webapps/cms]
  tags: [cms, open-source]
  aliases: [wp]
  icon: icons/wordpress.svg
  links:
    website: https://wordpress.org
    documentation: https://wordpress.org/documentation/
    source: https://github.com/WordPress/WordPress
  license: GPL-2.0-or-later
  requires:
    versionCode: v000001

parameters:
  - name: memoryLimit
    title: Memory limit
    type: size
    default: 512MB
    min: 256MB
  - name: dataVolume
    title: Site files
    description: The volume WordPress keeps its code, themes, plugins and uploads on.
    type: volume

dependencies:
  - name: db
    title: Database
    template: mysql
    version: "8.4"
    params:
      dbName: wordpress
      username: wordpress

versions:
  - name: "7.1"
    release: "7.1.0"
    default: true
    image: wordpress:7.1.0-php8.2-apache

app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        /var/www/html: {type: volume, source: "${{ params.dataVolume }}"}
    container:
      healthcheck:
        enabled: true
        mode: CMD-SHELL
        # Before installation this redirects to the installer, which curl -f accepts;
        # without a database WordPress answers 500, which it does not.
        command: curl -fsS -o /dev/null http://127.0.0.1/wp-login.php
        interval: 15s
        timeout: 5s
        startPeriod: 60s
        retries: 5
    resources:
      limits: {memory: "${{ params.memoryLimit }}"}
  settings:
    kind:
      category: webapp
      engine: wordpress
      version: "${{ version.name }}"
      webapp: {}
    envVars:
      data:
        - {k: WORDPRESS_DB_HOST, v: "${{ deps.db.ref.HIVEPAAS_HOST }}:${{ deps.db.ref.HIVEPAAS_PORT }}"}
        - {k: WORDPRESS_DB_USER, v: "${{ deps.db.ref.HIVEPAAS_USER }}"}
        - {k: WORDPRESS_DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
        - {k: WORDPRESS_DB_NAME, v: "${{ deps.db.ref.HIVEPAAS_DATABASE_NAME }}"}
    routing:
      port: 80
```

- [ ] **Step 4: Document dependencies in `README.md`**

After the section `## A template`, add:

````markdown
## Dependencies

A template can create other templates' apps alongside its own - a web app and its database:

```yaml
dependencies:
  - name: db              # the role, and the suffix of the app name: blog-db
    title: Database
    template: mysql       # a template in this repository
    version: "8.4"        # optional
    params:               # what this template fixes; the rest is defaulted or asked
      dbName: wordpress
```

The app refers to what a dependency shares with `${{ deps.db.ref.HIVEPAAS_PASSWORD }}`, which
renders to an ordinary reference, `${blog_db.HIVEPAAS_PASSWORD}`, resolved at deploy time.
`${{ deps.db.key }}` is the dependency app's key alone.

The rules, all checked by `apptemplate lint`:

- at most three dependencies, one level deep - a dependency cannot have dependencies;
- the named template, version, variant and parameters must exist in this repository;
- `ref.` names only what the dependency's kind shares - a database shares `HIVEPAAS_USER`,
  `HIVEPAAS_PASSWORD`, `HIVEPAAS_DATABASE_NAME` and `HIVEPAAS_SSL_MODE` besides the variables
  every app shares;
- `params` may use this template's parameters, but not its secrets.

A dependency parameter the template does not fix, with no default, not optional and not a
generated secret, is asked of the person creating the app - for a database, its data volume.
Deleting the app never deletes its dependencies.
````

- [ ] **Step 5: Lint and index**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
go run ./tools/apptemplate lint ../app-templates
go run ./tools/apptemplate index ../app-templates
go run ./tools/apptemplate index -check ../app-templates
```

Expected: `OK: 22 template(s)`, `written: ../app-templates/index.json`, `index.json is up to date`. Lint rendered WordPress bound to the MySQL template, which is the check that matters; `apptemplate render` renders one template on its own and is not used for a template with dependencies.

- [ ] **Step 6: Commit**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/app-templates
git add templates/wordpress.yaml icons/wordpress.svg tags.yaml README.md index.json
git commit -m "$(cat <<'EOF'
feat: add WordPress, the first template with a dependency

WordPress declares the MySQL template as its database. Creating it creates
both apps, and WordPress reaches the database through environment
references to what the database app shares, so the password is never
copied into its settings.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 12: End to end on a real backend

**Files:** none committed. Scripts live in the session scratchpad.

- [ ] **Step 1: Start the backend and prepare**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
HP_CONFIG_FILE=$PWD/config/config.local.toml HP_RUN_MODE=app+worker HP_HTTP_SERVER_PORT=10077 \
  HP_TEMPLATES_DIR=$PWD/../app-templates go run ./hivepaas_app/cmd/app > /tmp/hp-deps.log 2>&1 &
until curl -s -o /dev/null localhost:10077/_/ping; do sleep 3; done
```

Source the session's test helpers - `lib.sh` in the scratchpad, which defines `token`, `api`, `db`, `cid`, `wait_ready` and the development project's `PROJ=01JAB9XED0GTXBSQDFVYAJ8WB2`, `ENV=dev`, `NET=project_b_dev_net` - then create an inheritable volume and keep its id:

```bash
TOKEN=$(token)
VOL=$(api POST "/projects/$PROJ/$ENV/cluster-volumes" '{"name":"deps-vol","kind":"local","inheritable":true}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["id"])')
```

- [ ] **Step 2: Create WordPress with its database**

```bash
api POST "/projects/$PROJ/$ENV/apps/from-template" "$(python3 -c "
import json; print(json.dumps({'name': 'blog', 'template': 'wordpress',
  'params': {'dataVolume': '$VOL'}, 'dependencyParams': {'db': {'dataVolume': '$VOL'}}}))")" > /tmp/blog.json
BLOG_ID=$(python3 -c 'import json; print(json.load(open("/tmp/blog.json"))["data"]["app"]["id"])')
DB_ID=$(python3 -c 'import json; d=json.load(open("/tmp/blog.json"))["data"]["dependencies"][0]; print(d["app"]["id"])')
python3 -m json.tool /tmp/blog.json
```

Expected: `data.app.id`, `data.deployment.id`, and `data.dependencies[0]` with `name` `db`, its own app and its own deployment; `BLOG_ID` and `DB_ID` set.

- [ ] **Step 3: Both apps come up**

Run: `wait_ready blog_db` then `wait_ready blog`.
Expected: both return 0 - deployments `done`, containers `healthy`.

- [ ] **Step 4: WordPress reaches its database with the database's credentials**

```bash
[ "$(docker exec "$(cid blog)" printenv WORDPRESS_DB_PASSWORD)" = "$(docker exec "$(cid blog_db)" printenv MYSQL_PASSWORD)" ] && echo same
docker exec "$(cid blog)" printenv WORDPRESS_DB_HOST
docker run --rm --network "$NET" curlimages/curl:latest -s -o /dev/null -w '%{http_code}\n' http://blog/wp-admin/install.php
```

Expected: `same`; `blog_db:3306`; `200`.

- [ ] **Step 5: The bindings name each other**

```bash
api GET "/projects/$PROJ/$ENV/apps/$BLOG_ID/template"     # expect dependencies[0].appId == $DB_ID
api GET "/projects/$PROJ/$ENV/apps/$DB_ID/template"       # expect createdForAppId == $BLOG_ID
```

- [ ] **Step 6: Deleting WordPress leaves the database**

Delete `blog` through the API, wait for its containers to go, then:

```bash
db "SELECT key FROM apps WHERE key='blog_db' AND deleted_at IS NULL"
docker exec "$(cid blog_db)" sh -c 'mysql -u root -p"$MYSQL_ROOT_PASSWORD" -N -e "show databases like \"wordpress\"" 2>/dev/null'
```

Expected: `blog_db`; `wordpress`.

- [ ] **Step 7: A failed request leaves nothing behind**

Create an empty app named `blog2` (`POST /projects/$PROJ/$ENV/apps` with `{"name":"blog2","status":"active"}`), then create WordPress named `blog2` as in Step 2.

Expected: an error naming the conflict. Then:

```bash
db "SELECT count(*) FROM apps WHERE key='blog2_db' AND deleted_at IS NULL"
docker service ls --format '{{.Name}}' | grep -c 'project_b_dev_blog2_db'
```

Expected: `0` and `0` - the database provisioned first was rolled back and its service removed.

- [ ] **Step 8: Clean up**

Delete `blog_db`, `blog2` and the volume through the API, stop the backend, and confirm `docker ps` shows no `project_b_dev_blog*` container. Report the results of Steps 2 to 7 to the user with the exact outputs.

---

## Self-Review Notes

- Spec §2 decisions: one level (Tasks 2, 5, 6), at most three (Task 2), same repository and revision (Tasks 5, 6), always created (Task 8), never deleted (Task 11 description, Task 12 Step 6), asked parameters (Tasks 2, 4, 9), creation only (no update path touched).
- Spec §3 format and placeholders: Tasks 2 and 4. §4 provisioning and failure: Tasks 6 and 8, spec aligned in Task 10. §5 recording: Tasks 7 and 8. §6 index and store: Tasks 1, 5 and 9; the dashboard is out of scope per Global Constraints. §7 linting: Task 5. §8 security: Tasks 4 (secrets), 6 (revision, caps). §9 testing: every task's unit tests, Task 12 live.
- Names used across tasks: `DepBinding`, `DependencyParams`, `SharedVarsOf`, `KindCategory`, `AskedParams`, `DependencyAppName`, `DependencyTemplate`, `RenderedDependency`, `appToProvision`, `appTemplateLinks`, `planApps`, `provisionAll`, `removeServices`, `transformCreated`, `CreatedDependencyResp`, `AppTemplateDependencySummaryResp`, `AppTemplateDependencyResp`, `AppTemplateBindingDependencyResp` - each defined once, in the task listed in its Interfaces block.
