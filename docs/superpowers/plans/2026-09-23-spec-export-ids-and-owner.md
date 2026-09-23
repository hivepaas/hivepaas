# Spec Export: Ids and Project Owner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** An exported bundle carries the source id of every project, app and collection
setting, and each project's owner as `{id, email}`, so that a later import can match objects
exactly and find a project's owner again.

**Architecture:** Three additive, optional fields in the bundle format (`specmodel`), written by
the existing export walk (`specserviceimpl/walk.go`) and settings assembly
(`specserviceimpl/assemble.go`). Nothing reads them yet - import is a later plan. A guard test
keeps the `id` export writes into a collection entry from colliding with a field of the
setting's own data.

**Tech Stack:** Go 1.27, gopkg.in/yaml.v3, bun (relation loading), testify.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §4 *Matching*, "The export change" and "The project owner"; correction 1.

## Global Constraints

- Before calling a task done: `go build ./...`, `golangci-lint run ./...` over the whole repo
  (120-character lines, US spelling), and the tests the task names. The last task runs
  `go test ./...`.
- Every new field is optional in the format: `yaml:"...,omitempty"`. A bundle without them
  stays valid.
- `ProjectDoc` and `AppDoc` gain `id`. `EnvDoc` does not: an env's id is derived as
  `projecthelper.CalcProjectEnvID(projectID, name)`.
- A collection entry gains a top-level `id`. A singleton setting gains nothing: it is identified
  by `(scope, type)`.
- `ProjectDoc` gains `owner: { id, email }`. The email is not a secret, so it is written in every
  secrets mode, `omit` included.
- Two exports of unchanged data stay byte-identical apart from `exportedAt` -
  `TestExportIsDeterministicApartFromTheTimestamp` must keep passing untouched.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

## File Structure

| file | change |
|---|---|
| `hivepaas_app/service/specservice/specmodel/document.go` | `CollectionEntryIDKey`; `ID` on `ProjectDoc` and `AppDoc`; `ProjectOwner` and `ProjectDoc.Owner` |
| `hivepaas_app/service/specservice/specserviceimpl/assemble.go` | write the setting id into each collection entry |
| `hivepaas_app/service/specservice/specserviceimpl/walk.go` | write project and app ids; load and write the project owner |
| `hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go` | new: every test in this plan |
| `hivepaas_app/service/specservice/specserviceimpl/export_test.go` | the fixture's project gains an owner |

All commands run from `/Users/tnt/go/src/github.com/hivepaas/hivepaas`.

---

### Task 1: Collection entries carry their setting's id

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/document.go` (after `ExternalRef`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/assemble.go` (`assembleSettings`, the
  collection branch)
- Create: `hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go`

**Interfaces:**
- Consumes: `runExport(t, mode, passphrase) (string, *specmodel.Report)` and
  `readFromArchive(t, archive, name) string`, both in `export_test.go`.
- Produces: `specmodel.CollectionEntryIDKey = "id"`, which the import plan uses to take the id
  back out of an entry before decoding it. Test helpers `readDoc`, `entry` and `jsonKeys` in
  `export_ids_test.go`, which Tasks 2 and 3 use.

- [ ] **Step 1: Write the failing tests**

Create `hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go`:

```go
package specserviceimpl

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// readDoc parses one payload file of an exported archive into out.
func readDoc(t *testing.T, archive, name string, out any) {
	t.Helper()
	assert.NoError(t, yaml.Unmarshal([]byte(readFromArchive(t, archive, name)), out))
}

// entry is one entry of a collection block, as a document carries it.
func entry(t *testing.T, settings map[string]any, block, key string) map[string]any {
	t.Helper()
	entries, ok := settings[block].(map[string]any)
	if !assert.True(t, ok, "block %s is missing", block) {
		return nil
	}
	body, ok := entries[key].(map[string]any)
	assert.True(t, ok, "entry %s/%s is missing", block, key)
	return body
}

// A collection entry carries the id of the setting it was exported from, so an
// import into the same installation finds that setting even after a rename.
func TestExportWritesTheIDOfEveryCollectionEntry(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	global := &specmodel.GlobalDoc{}
	readDoc(t, path, "global.yaml", global)
	cert := entry(t, global.Settings, "sslCerts", "localhost")
	assert.Equal(t, "cert_1", cert[specmodel.CollectionEntryIDKey])

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	if assert.Contains(t, env.Apps, "backend") {
		secret := entry(t, env.Apps["backend"].Settings, "secrets", "db-password")
		assert.Equal(t, "secret_1", secret[specmodel.CollectionEntryIDKey])
	}
}

// A singleton is identified by its scope and its type, so it carries no id.
func TestExportWritesNoIDOnASingleton(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	if !assert.Contains(t, env.Apps, "backend") {
		return
	}
	routing, ok := env.Apps["backend"].Settings["routing"].(map[string]any)
	if assert.True(t, ok, "precondition: the fixture's app has a routing setting") {
		assert.NotContains(t, routing, specmodel.CollectionEntryIDKey)
	}
}

// Export writes a collection entry's id beside the setting's own data, at the
// top level of the entry. A setting type with a top-level field of that name
// would have the two collide - and encoding/json matches keys without regard to
// case, so a field called ID or Id collides as well.
func TestNoExportedCollectionTypeHasATopLevelID(t *testing.T) {
	checked := 0
	for _, typ := range entity.AllParsedSettingTypes() {
		if specmodel.CollectionBlockName(typ) == "" {
			continue
		}
		decision := entity.SpecExportDecision(&entity.Setting{
			Type: typ, Scope: base.ObjectScopeProject, ObjectID: "prj_1",
		})
		if !decision.Export {
			continue
		}
		data, err := (&entity.Setting{Type: typ}).Parse()
		if !assert.NoError(t, err, "%v", typ) {
			continue
		}
		checked++
		for _, key := range jsonKeys(reflect.TypeOf(data)) {
			assert.False(t, strings.EqualFold(key, specmodel.CollectionEntryIDKey),
				"%v has a top-level %q, which the id export writes would overwrite", typ, key)
		}
	}
	assert.Positive(t, checked, "precondition: some collection type is exported")
}

// jsonKeys lists the keys encoding/json writes at the top level of a value of
// typ: an untagged embedded struct is flattened into its parent, a field tagged
// "-" is left out, and an untagged field is written under its Go name.
func jsonKeys(typ reflect.Type) []string {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	var keys []string
	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if field.Anonymous && name == "" {
			keys = append(keys, jsonKeys(field.Type)...)
			continue
		}
		if !field.IsExported() {
			continue
		}
		if name == "" {
			name = field.Name
		}
		keys = append(keys, name)
	}
	return keys
}

// The guard above is only as good as jsonKeys, so jsonKeys is shown to find an
// id in each place encoding/json would put one, and not where it would not.
func TestJSONKeysFindsAnIDWhereverJSONWouldWriteOne(t *testing.T) {
	type embedded struct {
		ID string `json:"id"`
	}
	type withEmbedded struct {
		embedded

		Name string `json:"name"`
	}
	type untagged struct {
		ID string
	}
	type hidden struct {
		ID string `json:"-"`
	}

	// Built as values rather than named as types, so that every field is written
	// and the unused linter has nothing to report.
	assert.Contains(t, jsonKeys(reflect.TypeOf(withEmbedded{embedded: embedded{ID: "a"}, Name: "b"})), "id")
	assert.Contains(t, jsonKeys(reflect.TypeOf(&untagged{ID: "a"})), "ID")
	assert.Empty(t, jsonKeys(reflect.TypeOf(hidden{ID: "a"})))
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run 'TestExportWrites|TestNoExportedCollectionTypeHasATopLevelID|TestJSONKeys' -v`

Expected: the package does not compile - `undefined: specmodel.CollectionEntryIDKey`.

- [ ] **Step 3: Declare the key**

In `hivepaas_app/service/specservice/specmodel/document.go`, after the `ExternalRef` type:

```go
// CollectionEntryIDKey is where export writes the id of the setting a collection
// entry was exported from: at the top level of the entry, beside the setting's
// own data. An import into the installation that exported it matches on it, so
// a setting renamed since is still recognized. No exported collection type may
// have a top-level field of that name - TestNoExportedCollectionTypeHasATopLevelID
// holds that.
const CollectionEntryIDKey = "id"
```

- [ ] **Step 4: Write the id into each collection entry**

In `hivepaas_app/service/specservice/specserviceimpl/assemble.go`, in `assembleSettings`, the
collection branch becomes:

```go
		case specmodel.CollectionBlockName(typ) != "":
			keys := specmodel.DeriveSettingKeys(group)
			entries := map[string]any{}
			for _, setting := range group {
				body, err := renderSetting(setting, index, mode)
				if err != nil {
					return nil, hperrors.Wrap(err)
				}
				body[specmodel.CollectionEntryIDKey] = setting.ID
				entries[keys[setting.ID]] = body
			}
			out[specmodel.CollectionBlockName(typ)] = entries
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestExport|TestNoExportedCollectionTypeHasATopLevelID|TestJSONKeys'`

Expected: PASS, including every existing `TestExport*` - in particular
`TestExportIsDeterministicApartFromTheTimestamp` and `TestExportResolvesCrossScopeReferences`
(the env file now holds `secret_1` and `app` ids, never `cert_1`).

- [ ] **Step 6: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specmodel/document.go \
        hivepaas_app/service/specservice/specserviceimpl/assemble.go \
        hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go
git commit -F - <<'EOF'
feat(spec): write each collection setting's id into its entry

An import into the installation that exported a bundle matches a
collection setting on it, so a setting renamed since the export is still
recognized instead of created again. A test keeps any exported type from
having a top-level field the id would overwrite.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 2: Project and app documents carry their ids

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/document.go` (`ProjectDoc`, `AppDoc`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/walk.go` (`writeDocs`, `buildAppDoc`)
- Test: `hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go`

**Interfaces:**
- Consumes: `readDoc` from Task 1.
- Produces: `specmodel.ProjectDoc.ID string` and `specmodel.AppDoc.ID string`, both
  `yaml:"id,omitempty"`.

- [ ] **Step 1: Write the failing test**

Append to `export_ids_test.go`:

```go
// Projects and apps carry their ids too. An env does not need one: its id is
// derived from its project's id and its own name.
func TestExportWritesTheIDsOfProjectsAndApps(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	project := &specmodel.ProjectDoc{}
	readDoc(t, path, "projects/project_a/project.yaml", project)
	assert.Equal(t, "p1", project.ID)

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	if assert.Contains(t, env.Apps, "backend") && assert.Contains(t, env.Apps, "frontend") {
		assert.Equal(t, "app_1", env.Apps["backend"].ID)
		assert.Equal(t, "app_2", env.Apps["frontend"].ID, "an app never deployed carries its id as well")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestExportWritesTheIDsOfProjectsAndApps -v`

Expected: the package does not compile - `project.ID undefined (type *specmodel.ProjectDoc has
no field or method ID)`.

- [ ] **Step 3: Add the fields**

In `hivepaas_app/service/specservice/specmodel/document.go`, `ProjectDoc` becomes:

```go
// ProjectDoc is projects/<key>/project.yaml.
type ProjectDoc struct {
	DocHeader `yaml:",inline"`

	Project string `yaml:"project"`
	// ID is the project's id on the installation that exported it. An import
	// into that installation matches on it; anywhere else it matches nothing,
	// and the key is used instead.
	ID   string `yaml:"id,omitempty"`
	Name string `yaml:"name"`
	Note string `yaml:"note,omitempty"`
	// Envs names the env files belonging to this project, so a reader of one
	// project file knows what else there is without listing the archive.
	Envs []string `yaml:"envs,omitempty"`

	Settings map[string]any `yaml:"settings,omitempty"`
}
```

and `AppDoc` becomes:

```go
type AppDoc struct {
	App string `yaml:"app"`
	// ID is the app's id on the installation that exported it; see ProjectDoc.ID.
	ID     string `yaml:"id,omitempty"`
	Name   string `yaml:"name"`
	Status string `yaml:"status,omitempty"`
	Note   string `yaml:"note,omitempty"`

	Deployment *Deployment `yaml:"deployment,omitempty"`

	Settings map[string]any `yaml:"settings,omitempty"`
}
```

Keep `AppDoc`'s existing doc comment above it unchanged. A template never sets `id`, and
`CheckBuildable` refuses one that does: `onlyFields("", doc, "deployment", "settings")` reports
any non-zero field outside those two.

- [ ] **Step 4: Write them**

In `hivepaas_app/service/specservice/specserviceimpl/walk.go`, in `writeDocs`, the project
document gains its id:

```go
		doc := &specmodel.ProjectDoc{
			DocHeader: specmodel.NewDocHeader(specScopeName(base.ObjectScopeProject)),
			Project:   projUnit.project.Key,
			ID:        projUnit.project.ID,
			Name:      projUnit.project.Name,
			Note:      projUnit.project.Note,
			Envs:      envNames,
			Settings:  assembled,
		}
```

and in `buildAppDoc`:

```go
	doc := &specmodel.AppDoc{
		App:      app.app.Key,
		ID:       app.app.ID,
		Name:     app.app.Name,
		Status:   string(app.app.Status),
		Note:     app.app.Note,
		Settings: assembled,
	}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestExport|TestBuild|TestCheckBuildable'`

Expected: PASS - the builder and `CheckBuildable` tests included, since `AppDoc` gained a field.

- [ ] **Step 6: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specmodel/document.go \
        hivepaas_app/service/specservice/specserviceimpl/walk.go \
        hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go
git commit -F - <<'EOF'
feat(spec): write the id of each exported project and app

An import into the installation that exported a bundle matches projects
and apps on their ids, which is also what tells the same app apart from
another one created later under the same key. Env ids are derived from
the project's id and the env's name, so envs need none.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 3: Project documents carry their owner

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/document.go` (`ProjectOwner`,
  `ProjectDoc.Owner`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/walk.go` (`projectsInScope`,
  `writeDocs`, new `projectOwner`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/export_test.go` (`exportFixture`)
- Test: `hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go`

**Interfaces:**
- Consumes: `readDoc` from Task 1; `ProjectDoc` as Task 2 left it.
- Produces: `specmodel.ProjectOwner{ID, Email string}` (both `omitempty`),
  `specmodel.ProjectDoc.Owner *ProjectOwner` (`yaml:"owner,omitempty"`), and
  `projectOwner(*entity.Project) *specmodel.ProjectOwner` in `specserviceimpl`.

- [ ] **Step 1: Give the fixture's project an owner**

In `hivepaas_app/service/specservice/specserviceimpl/export_test.go`, in `exportFixture`, the
exportable project becomes:

```go
	proj := &entity.Project{
		ID: "p1", Key: "project_a", Name: "Project A",
		OwnerID: "u1", Owner: &entity.User{ID: "u1", Email: "owner@example.com"},
	}
```

- [ ] **Step 2: Write the failing tests**

Append to `export_ids_test.go`:

```go
// Users do not travel in a bundle, so a project's owner is written as what
// finds them again: the id on the installation that exported it, the email
// anywhere else. The email is not a secret, and omit mode writes it too.
func TestExportWritesTheOwnerOfAProject(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	project := &specmodel.ProjectDoc{}
	readDoc(t, path, "projects/project_a/project.yaml", project)
	assert.Equal(t, &specmodel.ProjectOwner{ID: "u1", Email: "owner@example.com"}, project.Owner)
}

func TestProjectOwner(t *testing.T) {
	cases := map[string]struct {
		project *entity.Project
		want    *specmodel.ProjectOwner
	}{
		"no owner": {project: &entity.Project{}, want: nil},
		"owner loaded": {
			project: &entity.Project{OwnerID: "u1", Owner: &entity.User{ID: "u1", Email: "owner@example.com"}},
			want:    &specmodel.ProjectOwner{ID: "u1", Email: "owner@example.com"},
		},
		// The relation comes back empty when the user row is gone. The id is
		// written anyway: it records whose the project was, and an import that
		// finds no such user falls through to the operator importing.
		"owner row missing": {
			project: &entity.Project{OwnerID: "u1"},
			want:    &specmodel.ProjectOwner{ID: "u1"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, projectOwner(tc.project))
		})
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run 'TestExportWritesTheOwnerOfAProject|TestProjectOwner' -v`

Expected: the package does not compile - `undefined: specmodel.ProjectOwner` and
`undefined: projectOwner`.

- [ ] **Step 4: Add the owner to the format**

In `hivepaas_app/service/specservice/specmodel/document.go`, `ProjectDoc` gains `Owner` after
`Note`:

```go
	Name string `yaml:"name"`
	Note string `yaml:"note,omitempty"`
	// Owner is the user who owns the project, as import can find them again.
	Owner *ProjectOwner `yaml:"owner,omitempty"`
```

and below `ProjectDoc`:

```go
// ProjectOwner names the user who owns a project. Users do not travel in a
// bundle, so import finds the owner again: by id on the installation that
// exported it - which still works after the owner changed their email - and by
// email anywhere else.
type ProjectOwner struct {
	ID    string `yaml:"id,omitempty"`
	Email string `yaml:"email,omitempty"`
}
```

- [ ] **Step 5: Load the owner with the projects, and write it**

In `hivepaas_app/service/specservice/specserviceimpl/walk.go`, `projectsInScope` loads the owner
relation in both of its queries - the same option `projectuc.GetProject` uses:

```go
	case base.ObjectScopeGlobal:
		projects, _, err := s.projectRepo.List(ctx, db, nil, withProjectOwner())
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return selectProjects(projects), nil

	case base.ObjectScopeProject, base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		project, err := s.projectRepo.GetByID(ctx, db, scope.ProjectID, withProjectOwner())
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return selectProjects([]*entity.Project{project}), nil
```

Below `projectsInScope`, add:

```go
// withProjectOwner loads the user who owns each project, for the email a bundle
// carries. It excludes the columns every other reader of a user excludes.
func withProjectOwner() bunex.SelectQueryOption {
	return bunex.SelectRelation("Owner",
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
}

// projectOwner is a project's owner as a bundle carries it: the id always, and
// the email when the user row came back with the project. A project whose
// owner's row is gone carries the id alone.
func projectOwner(project *entity.Project) *specmodel.ProjectOwner {
	if project.OwnerID == "" {
		return nil
	}
	owner := &specmodel.ProjectOwner{ID: project.OwnerID}
	if project.Owner != nil {
		owner.Email = project.Owner.Email
	}
	return owner
}
```

In `writeDocs`, the project document gains `Owner`:

```go
		doc := &specmodel.ProjectDoc{
			DocHeader: specmodel.NewDocHeader(specScopeName(base.ObjectScopeProject)),
			Project:   projUnit.project.Key,
			ID:        projUnit.project.ID,
			Name:      projUnit.project.Name,
			Note:      projUnit.project.Note,
			Owner:     projectOwner(projUnit.project),
			Envs:      envNames,
			Settings:  assembled,
		}
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestExport|TestProjectOwner'`

Expected: PASS. The fakes ignore query options, so these tests cannot see the relation being
asked for; Task 4 checks that part against a real database.

- [ ] **Step 7: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specmodel/document.go \
        hivepaas_app/service/specservice/specserviceimpl/walk.go \
        hivepaas_app/service/specservice/specserviceimpl/export_test.go \
        hivepaas_app/service/specservice/specserviceimpl/export_ids_test.go
git commit -F - <<'EOF'
feat(spec): write each project's owner as an id and an email

Users do not travel in a bundle, so import will find the owner again by
id on the installation that exported it and by email anywhere else. The
email is not a secret, and omit mode writes it too.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 4: Verify the whole change

**Files:** none changed, unless a check fails.

- [ ] **Step 1: The whole suite**

Run: `go build ./... && golangci-lint run ./... && go test ./...`

Expected: everything passes.

- [ ] **Step 2: Export through a running backend**

The owner relation is the one part no unit test reaches, because the fakes ignore query
options. Run a development backend against a database that has at least one project, export
it from the dashboard - Operations → Export, secrets mode *Omit* - and unpack the download:

The download is named `hivepaas-spec-<timestamp>.tar.gz`; this unpacks the newest one:

```bash
BUNDLE=$(ls -t ~/Downloads/hivepaas-spec-*.tar.gz | head -1)
rm -rf /tmp/spec-check && mkdir -p /tmp/spec-check && tar -xzf "$BUNDLE" -C /tmp/spec-check
cat /tmp/spec-check/projects/*/project.yaml
grep -n "id:" /tmp/spec-check/projects/*/envs/*.yaml | head -20
```

Expected: each `project.yaml` has `id:` and an `owner:` block with the owner's `id` and
`email`; each app in an env file has `id:`; each entry of a collection block such as
`secrets:` or `configFiles:` has `id:`; `routing:`, `kind:` and `envVars:` have none.

- [ ] **Step 3: Report**

Say which checks ran and what they showed. If Step 2 could not be run, say so rather than
calling the task done.
