package specserviceimpl

import (
	"context"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	backendPath = "projects/project_a/envs/dev/apps/backend"
)

// planFixture is the export fixture's installation and a bundle it exported,
// read back, for a test to change before planning it.
func planFixture(t *testing.T) (*service, *specmodel.ImportBundle) {
	t.Helper()
	bundle, err := readBundle(exportedBytes(t, specmodel.SecretsModeOmit, ""), "")
	assert.NoError(t, err)
	svc, _ := exportFixture(t).(*service)
	return svc, bundle
}

func plan(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, options specmodel.ImportOptions,
) *specmodel.ImportPlan {
	t.Helper()
	if options.Existing == "" {
		options.Existing = specmodel.ExistingUpdate
	}
	out, err := svc.planImport(context.Background(), nil, &specservice.ValidateImportReq{
		Scope: entity.NewObjectScopeGlobal(), Options: options,
	}, bundle)
	assert.NoError(t, err)
	return out
}

func node(t *testing.T, p *specmodel.ImportPlan, path string) *specmodel.PlanNode {
	t.Helper()
	for _, n := range p.Nodes {
		if n.Path == path {
			return n
		}
	}
	t.Fatalf("no node %s", path)
	return nil
}

func backendRouting(bundle *specmodel.ImportBundle) map[string]any {
	routing, _ := bundle.Envs["project_a"]["dev"].Apps["backend"].Settings["routing"].(map[string]any)
	return routing
}

func globalCerts(bundle *specmodel.ImportBundle) map[string]any {
	certs, _ := bundle.Global.Settings["sslCerts"].(map[string]any)
	return certs
}

// A bundle imported into the installation that exported it changes nothing:
// every node is unchanged and nothing restarts. This is what makes running the
// same import twice safe.
func TestPlanOfAnUnchangedInstallationChangesNothing(t *testing.T) {
	svc, bundle := planFixture(t)

	p := plan(t, svc, bundle, specmodel.ImportOptions{})

	for _, n := range p.Nodes {
		assert.Equal(t, specmodel.ActionUnchanged, n.Action, "%s changed: %v", n.Path, n.Changes)
		assert.False(t, n.Restart, n.Path)
		assert.Empty(t, n.Issues, n.Path)
	}
	assert.Equal(t, specmodel.MatchedByID, node(t, p, backendPath).MatchedBy)
}

// A changed setting of a running app updates it, and says it will restart.
func TestPlanOfAChangedSettingRestartsTheApp(t *testing.T) {
	svc, bundle := planFixture(t)
	backendRouting(bundle)["port"] = 9090

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	assert.Equal(t, specmodel.ActionUpdate, backend.Action)
	assert.Equal(t, []string{"settings.routing"}, backend.Changes)
	assert.True(t, backend.Restart)
}

func TestPlanOfANewAppCreatesAndDeploysIt(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Envs["project_a"]["dev"].Apps["worker"] = &specmodel.AppDoc{
		App: "worker", Name: "Worker",
		Deployment: &specmodel.Deployment{Source: map[string]any{"activeMethod": "image"}},
	}

	worker := node(t, plan(t, svc, bundle, specmodel.ImportOptions{DeployCreated: true}),
		"projects/project_a/envs/dev/apps/worker")

	assert.Equal(t, specmodel.ActionCreate, worker.Action)
	assert.True(t, worker.Deploy)
}

// Keep leaves what exists alone, however it differs.
func TestPlanKeepLeavesExistingObjectsAlone(t *testing.T) {
	svc, bundle := planFixture(t)
	backendRouting(bundle)["port"] = 9090

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{Existing: specmodel.ExistingKeep}), backendPath)

	assert.Equal(t, specmodel.ActionKeep, backend.Action)
	assert.False(t, backend.Restart)
}

// A project whose id matches one here under another key was edited in the
// bundle; following the id would rename the original, so it is skipped with
// everything in it.
func TestPlanSkipsAProjectWhoseKeyWasEdited(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Projects["project_copy"] = bundle.Projects["project_a"]
	bundle.Envs["project_copy"] = bundle.Envs["project_a"]
	delete(bundle.Projects, "project_a")
	delete(bundle.Envs, "project_a")

	p := plan(t, svc, bundle, specmodel.ImportOptions{})

	project := node(t, p, "projects/project_copy")
	assert.Equal(t, specmodel.ActionSkip, project.Action)
	if assert.NotEmpty(t, project.Issues) {
		assert.Equal(t, specmodel.CodeKeyMismatch, project.Issues[0].Code)
	}
	assert.Equal(t, specmodel.ActionSkip, node(t, p, "projects/project_copy/envs/dev/apps/backend").Action)
}

func TestPlanRaisesIssuesASettingCarries(t *testing.T) {
	svc, bundle := planFixture(t)
	meta, _ := backendRouting(bundle)[specmodel.SettingMetaKey].(map[string]any)
	meta["version"] = 999
	bundle.Envs["project_a"]["dev"].Apps["backend"].Settings["mystery"] = map[string]any{"a": 1}

	backend := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), backendPath)

	codes := map[string]specmodel.Severity{}
	for _, issue := range backend.Issues {
		codes[issue.Code] = issue.Severity
	}
	assert.Equal(t, specmodel.SeverityBlocked, codes[specmodel.CodeSettingVersionNewer])
	assert.Equal(t, specmodel.SeveritySkipped, codes[specmodel.CodeTypeNotImportable])
}

// The hash binds apply to what the operator saw: the same request gives the same
// hash, and a different option a different one.
func TestPlanHash(t *testing.T) {
	// One upload: two exports differ in their timestamp, and so in their digest.
	content := exportedBytes(t, specmodel.SecretsModeOmit, "")
	planOf := func(options specmodel.ImportOptions) *specmodel.ImportPlan {
		bundle, err := readBundle(content, "")
		assert.NoError(t, err)
		svc, _ := exportFixture(t).(*service)
		return plan(t, svc, bundle, options)
	}
	first := planOf(specmodel.ImportOptions{})
	second := planOf(specmodel.ImportOptions{})
	other := planOf(specmodel.ImportOptions{Existing: specmodel.ExistingKeep})

	assert.Len(t, first.PlanHash, 64)
	assert.Equal(t, first.PlanHash, second.PlanHash)
	assert.NotEqual(t, first.PlanHash, other.PlanHash)
}

// A project route plans its own project, found in the bundle by id, and
// nothing else in the bundle.
func TestPlanAtAProjectRouteTakesThatProjectOnly(t *testing.T) {
	svc, bundle := planFixture(t)

	out, err := svc.planImport(context.Background(), nil, &specservice.ValidateImportReq{
		Scope: entity.NewObjectScopeProject("p1"), Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate},
	}, bundle)

	assert.NoError(t, err)
	for _, n := range out.Nodes {
		assert.NotEqual(t, "global", n.Path)
	}
	assert.Equal(t, specmodel.ActionUnchanged, node(t, out, backendPath).Action)
}

// Keep still creates what the target does not have at all: a certificate is an
// object of its own, and adding it changes nothing that exists.
func TestPlanKeepCreatesTheSettingsAScopeLacks(t *testing.T) {
	svc, bundle := planFixture(t)
	certs := globalCerts(bundle)
	fresh := maps.Clone(certs["localhost"].(map[string]any))
	fresh[specmodel.CollectionEntryIDKey] = "cert_new"
	certs["fresh"] = fresh
	certs["localhost"].(map[string]any)["domain"] = "changed.example.com"

	global := node(t, plan(t, svc, bundle, specmodel.ImportOptions{Existing: specmodel.ExistingKeep}), "global")

	assert.Equal(t, specmodel.ActionUpdate, global.Action)
	assert.Equal(t, []string{"sslCerts/fresh"}, global.Changes)
}

// A setting renamed since the export is the setting the target has, matched by
// its id: changed, and not missing.
func TestPlanMatchesARenamedSettingByID(t *testing.T) {
	svc, bundle := planFixture(t)
	certs := globalCerts(bundle)
	renamed := certs["localhost"].(map[string]any)
	renamed[specmodel.SettingMetaKey].(map[string]any)["name"] = "renamed"
	certs["renamed"] = renamed
	delete(certs, "localhost")

	updated := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), "global")
	kept := node(t, plan(t, svc, bundle, specmodel.ImportOptions{Existing: specmodel.ExistingKeep}), "global")

	assert.Equal(t, specmodel.ActionUpdate, updated.Action)
	assert.Equal(t, []string{"sslCerts/renamed"}, updated.Changes)
	assert.Equal(t, specmodel.ActionKeep, kept.Action)
}
