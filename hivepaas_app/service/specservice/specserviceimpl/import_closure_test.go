package specserviceimpl

import (
	"context"
	"os"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// replaceCert gives the bundle a certificate the target does not have - another
// key, another setting - with the backend's domain on it. name is what the
// certificate is called, which is what finds it by name.
func replaceCert(bundle *specmodel.ImportBundle, key, name string) {
	certs := globalCerts(bundle)
	cert, _ := certs["localhost"].(map[string]any)
	delete(certs, "localhost")
	cert[specmodel.CollectionEntryIDKey] = "cert_new"
	meta, _ := cert[specmodel.SettingMetaKey].(map[string]any)
	meta[settingMetaName] = name
	certs[key] = cert

	domains, _ := backendRouting(bundle)["domains"].([]any)
	domain, _ := domains[0].(map[string]any)
	domain["sslCert"] = map[string]any{"id": "global/sslCerts/" + key}
}

func backendStorage(bundle *specmodel.ImportBundle) *specmodel.Storage {
	return bundle.Envs["project_a"]["dev"].Apps["backend"].Deployment.Storage
}

func codesOf(node *specmodel.PlanNode) map[string]specmodel.Issue {
	out := map[string]specmodel.Issue{}
	for _, issue := range node.Issues {
		out[issue.Code] = issue
	}
	return out
}

// Selecting one app pulls in the certificate it uses when the target lacks it -
// that one setting, not the rest of global.
func TestPlanPullsInASettingTheTargetLacks(t *testing.T) {
	svc, bundle := planFixture(t)
	replaceCert(bundle, "fresh", "fresh")

	p := plan(t, svc, bundle, specmodel.ImportOptions{}, backendPath)

	global := node(t, p, "global")
	assert.True(t, global.Selected)
	assert.Equal(t, selectedByDependency, global.SelectedBy)
	assert.Equal(t, specmodel.ActionUpdate, global.Action)
	assert.Equal(t, []string{"sslCerts/fresh"}, global.Changes)
	assert.Empty(t, node(t, p, backendPath).Issues)
	assert.False(t, node(t, p, "projects/project_a/settings").Selected,
		"the volume the target has stays where it is")
}

// A dependency the operator deselected is looked for on the target, by its
// name; failing that, the reference is cleared.
func TestPlanClearsAReferenceToWhatTheOperatorExcluded(t *testing.T) {
	svc, bundle := planFixture(t)
	replaceCert(bundle, "fresh", "fresh")

	p := plan(t, svc, bundle, specmodel.ImportOptions{}, backendPath, "-global")

	assert.False(t, node(t, p, "global").Selected)
	issue, found := codesOf(node(t, p, backendPath))[specmodel.CodeRefNotSelected]
	if assert.True(t, found) {
		assert.Equal(t, specmodel.SeverityFixable, issue.Severity)
		assert.Equal(t, "global", issue.AvailableIn)
		assert.Equal(t, map[string]any{"setting": "settings.routing", "ref": "global/sslCerts/fresh"}, issue.Detail)
	}
}

func TestPlanResolvesAnExcludedDependencyByName(t *testing.T) {
	svc, bundle := planFixture(t)
	replaceCert(bundle, "renamed", "localhost")

	p := plan(t, svc, bundle, specmodel.ImportOptions{}, backendPath, "-global")

	assert.Empty(t, node(t, p, backendPath).Issues, "the target's localhost certificate is the one")
}

// A volume nobody has, a certificate export could not resolve, and an app whose
// directory a mount reaches that neither side has.
func TestPlanReportsReferencesNothingSatisfies(t *testing.T) {
	svc, bundle := planFixture(t)
	storage := backendStorage(bundle)
	storage.Mounts["/shared"] = specmodel.Mount{Type: mount.TypeVolume,
		External: &specmodel.ExternalRef{Type: "cluster-volume", Name: "nowhere"}}
	data := storage.Mounts["/var/lib/postgresql/data"]
	data.SourceApp = &specmodel.MountSourceApp{App: "ghost"}
	storage.Mounts["/var/lib/postgresql/data"] = data
	domains, _ := backendRouting(bundle)["domains"].([]any)
	domains[0].(map[string]any)["sslCert"] = map[string]any{"id": "01GONE"}

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	var details []map[string]any
	for _, issue := range backend.Issues {
		assert.Equal(t, specmodel.CodeRefNotFound, issue.Code)
		details = append(details, issue.Detail)
	}
	assert.ElementsMatch(t, []map[string]any{
		{"mount": "/shared", "ref": map[string]any{"type": "cluster-volume", "name": "nowhere"}},
		{"mount": "/var/lib/postgresql/data", "ref": "ghost"},
		{"setting": "settings.routing", "ref": "01GONE"},
	}, details)
}

// A mount reaching the directory of an app the env has here needs nothing.
func TestPlanResolvesASourceAppTheTargetHas(t *testing.T) {
	svc, bundle := planFixture(t)
	storage := backendStorage(bundle)
	data := storage.Mounts["/var/lib/postgresql/data"]
	data.SourceApp = &specmodel.MountSourceApp{App: "frontend"}
	storage.Mounts["/var/lib/postgresql/data"] = data

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	assert.Equal(t, []string{"deployment.storage"}, backend.Changes)
	assert.Empty(t, backend.Issues)
}

// An env route plans a project's bundle without the project's own settings; the
// volume the app mounts is found on the target, and so is the certificate.
func TestPlanAtAnEnvRouteResolvesTheProjectsSettingsOnTheTarget(t *testing.T) {
	path, _ := runExportAt(t, entity.NewObjectScopeProject("p1"), specmodel.SecretsModeOmit, "")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	bundle, err := readBundle(content, "")
	assert.NoError(t, err)
	svc, _ := exportFixture(t).(*service)
	storage := backendStorage(bundle)
	data := storage.Mounts["/var/lib/postgresql/data"]
	data.VolumeOptions = &specmodel.VolumeOptions{Subpath: "data2"}
	storage.Mounts["/var/lib/postgresql/data"] = data
	backendRouting(bundle)["port"] = 9090

	out, err := svc.planImport(context.Background(), nil, &specservice.ValidateImportReq{
		Scope:   entity.NewObjectScopeProjectEnv("p1", "dev"),
		Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate},
	}, bundle)

	assert.NoError(t, err)
	backend := node(t, out, backendPath)
	assert.ElementsMatch(t, []string{"deployment.storage", "settings.routing"}, backend.Changes)
	assert.Empty(t, backend.Issues)
}

// An app selected in a project the target lacks brings its project's and env's
// records, and not their settings.
func TestPlanCreatesTheRecordsASelectedAppNeeds(t *testing.T) {
	svc, bundle := planFixture(t)
	other, err := readBundle(exportedBytes(t, specmodel.SecretsModeOmit, ""), "")
	assert.NoError(t, err)
	project := other.Projects["project_a"]
	project.ID, project.Name = "", "Project New"
	for _, app := range other.Envs["project_a"]["dev"].Apps {
		app.ID = ""
	}
	bundle.Projects["project_new"] = project
	bundle.Envs["project_new"] = other.Envs["project_a"]

	p := plan(t, svc, bundle, specmodel.ImportOptions{}, "projects/project_new/envs/dev/apps/frontend")

	for _, path := range []string{"projects/project_new", "projects/project_new/envs/dev"} {
		record := node(t, p, path)
		assert.True(t, record.Selected, path)
		assert.Equal(t, selectedByDependency, record.SelectedBy, path)
		assert.Equal(t, specmodel.ActionCreate, record.Action, path)
	}
	assert.False(t, node(t, p, "projects/project_new/settings").Selected)
	assert.False(t, node(t, p, "projects/project_new/envs/dev/settings").Selected)
	assert.False(t, node(t, p, "projects/project_new/envs/dev/apps/backend").Selected)
}
