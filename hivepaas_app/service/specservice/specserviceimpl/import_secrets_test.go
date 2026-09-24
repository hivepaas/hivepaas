package specserviceimpl

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const bundlePassphrase = "correct horse battery staple"

// encryptedPlanFixture is planFixture with a bundle that carries its secrets.
func encryptedPlanFixture(t *testing.T) (*service, *specmodel.ImportBundle) {
	t.Helper()
	bundle, err := readBundle(exportedBytes(t, specmodel.SecretsModeEncrypted, bundlePassphrase), bundlePassphrase)
	assert.NoError(t, err)
	svc, _ := exportFixture(t).(*service)
	return svc, bundle
}

func backendSettings(bundle *specmodel.ImportBundle) map[string]any {
	return bundle.Envs["project_a"]["dev"].Apps["backend"].Settings
}

func notesOf(node *specmodel.PlanNode, code string) []specmodel.Issue {
	var out []specmodel.Issue
	for _, note := range node.Notes {
		if note.Code == code {
			out = append(out, note)
		}
	}
	return out
}

// A bundle without its secrets creates a secret empty. The secret the target
// already has keeps its value, and raises nothing.
func TestPlanOfAnOmitBundleLeavesACreatedSecretEmpty(t *testing.T) {
	svc, bundle := planFixture(t)
	secrets, _ := backendSettings(bundle)["secrets"].(map[string]any)
	secrets["NEW_TOKEN"] = map[string]any{"key": "NEW_TOKEN", "value": ""}

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	if issues := issuesOf(backend, specmodel.CodeSecretOmitted); assert.Len(t, issues, 1) {
		assert.Equal(t, specmodel.SeverityFixable, issues[0].Severity)
		assert.Equal(t, map[string]any{"setting": "settings.secrets/NEW_TOKEN", "secrets": 1}, issues[0].Detail)
	}
}

// A credential HivePaaS owns is generated for an app being created, which
// initializes its storage with it.
func TestPlanOfAnOmitBundleGeneratesACreatedAppsCredential(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, map[string]any{"kind": map[string]any{
		"category": "database", "engine": "postgres", "database": map[string]any{"dbName": "worker"},
	}})

	worker := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), workerPath)

	assert.Empty(t, issuesOf(worker, specmodel.CodeSecretOmitted))
	if notes := notesOf(worker, specmodel.CodeSecretGenerated); assert.Len(t, notes, 1) {
		assert.Equal(t, map[string]any{"setting": "settings.kind"}, notes[0].Detail)
	}
}

// An app matched by key is not the app the bundle's credential was made for:
// its data was initialized with the target's, which it keeps. The credential
// is then no difference.
func TestPlanOfAnEncryptedBundleKeepsTheCredentialOfAnAppMatchedByKey(t *testing.T) {
	svc, bundle := encryptedPlanFixture(t)
	bundle.Envs["project_a"]["dev"].Apps["backend"].ID = "app_elsewhere"
	kind, _ := backendSettings(bundle)["kind"].(map[string]any)
	kind["database"].(map[string]any)["password"] = "another"
	backendRouting(bundle)["port"] = 9090

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	assert.Equal(t, specmodel.MatchedByKey, backend.MatchedBy)
	assert.Equal(t, []string{"settings.routing"}, backend.Changes)
	if notes := notesOf(backend, specmodel.CodeCredentialKept); assert.Len(t, notes, 1) {
		assert.Equal(t, map[string]any{"setting": "settings.kind"}, notes[0].Detail)
	}
}

// The same app, matched by id, takes the bundle's credential.
func TestPlanOfAnEncryptedBundleWritesTheCredentialOfAnAppMatchedByID(t *testing.T) {
	svc, bundle := encryptedPlanFixture(t)
	kind, _ := backendSettings(bundle)["kind"].(map[string]any)
	kind["database"].(map[string]any)["password"] = "another"

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	assert.Equal(t, specmodel.MatchedByID, backend.MatchedBy)
	assert.Equal(t, []string{"settings.kind"}, backend.Changes)
	assert.Empty(t, backend.Notes)
}

// appKindCredentials names every secret of the app-kind type, and nothing else:
// a credential added to the type and missed here would be overwritten by an
// import it should have survived.
func TestAppKindCredentialsAreItsSecrets(t *testing.T) {
	jsonName := func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		return name
	}
	var secrets [][]string
	kind := reflect.TypeFor[entity.AppKindSettings]()
	for i := range kind.NumField() {
		block := kind.Field(i)
		if block.Type.Kind() != reflect.Pointer || block.Type.Elem().Kind() != reflect.Struct {
			continue
		}
		for j := range block.Type.Elem().NumField() {
			field := block.Type.Elem().Field(j)
			if field.Type == reflect.TypeFor[entity.EncryptedField]() {
				secrets = append(secrets, []string{jsonName(block), jsonName(field)})
			}
		}
	}

	assert.ElementsMatch(t, secrets, appKindCredentials)
}
