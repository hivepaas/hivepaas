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

// A setting is more than its data: the row says what it is called, what kind it
// is, whether it is on, whether the scopes below see it, whether it is the
// default of its kind, and which version of its data it holds.
func TestExportWritesTheRowOfEverySetting(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	global := &specmodel.GlobalDoc{}
	readDoc(t, path, "global.yaml", global)
	cert := entry(t, global.Settings, "sslCerts", "localhost")
	assert.Equal(t, map[string]any{"name": "localhost", "kind": "self-signed", "status": "active", "version": 1},
		cert[specmodel.SettingMetaKey])

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing, ok := env.Apps["backend"].Settings["routing"].(map[string]any)
	if assert.True(t, ok) {
		assert.Equal(t, map[string]any{"status": "active", "version": 1}, routing[specmodel.SettingMetaKey],
			"a singleton has no name, and the row still says the rest")
	}
}

func TestNoExportedSettingTypeHasATopLevelSettingField(t *testing.T) {
	for _, typ := range entity.AllParsedSettingTypes() {
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
		for _, key := range jsonKeys(reflect.TypeOf(data)) {
			assert.False(t, strings.EqualFold(key, specmodel.SettingMetaKey),
				"%v has a top-level %q, which the row export writes would overwrite", typ, key)
		}
	}
}
