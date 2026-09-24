package specserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// applyReq is a global apply of what selectors choose, with the hash of the
// plan validate makes of it, and every issue accepted.
func applyReq(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, selectors ...string,
) *specservice.ApplyImportReq {
	t.Helper()
	return applyReqWith(t, svc, bundle, specmodel.ImportOptions{}, selectors...)
}

func applyReqWith(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, options specmodel.ImportOptions, selectors ...string,
) *specservice.ApplyImportReq {
	t.Helper()
	validated := plan(t, svc, bundle, options, selectors...)
	req := &specservice.ApplyImportReq{OperatorID: "u_operator", PlanHash: validated.PlanHash, AcceptIssues: true}
	req.Scope = entity.NewObjectScopeGlobal()
	req.Options = options
	if req.Options.Existing == "" {
		req.Options.Existing = specmodel.ExistingUpdate
	}
	for _, selector := range selectors {
		if excluded, ok := cutExclude(selector); ok {
			req.Selection.Exclude = append(req.Selection.Exclude, excluded)
		} else {
			req.Selection.Include = append(req.Selection.Include, selector)
		}
	}
	return req
}

func cutExclude(selector string) (string, bool) {
	if len(selector) > 0 && selector[0] == '-' {
		return selector[1:], true
	}
	return selector, false
}

func apply(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, req *specservice.ApplyImportReq,
) *specservice.ApplyImportResp {
	t.Helper()
	resp, err := svc.applyBundle(context.Background(), nil, req, bundle)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return resp
}

func persisted(svc *service) *projectservice.PersistingProjectData {
	data := svc.projectService.(*fakeProjectService).persisted
	return data
}

func persistedSetting(t *testing.T, svc *service, match func(*entity.Setting) bool) *entity.Setting {
	t.Helper()
	for _, setting := range persisted(svc).UpsertingSettings {
		if match(setting) {
			return setting
		}
	}
	t.Fatalf("no such setting was written")
	return nil
}

func ofType(typ base.SettingType) func(*entity.Setting) bool {
	return func(s *entity.Setting) bool { return s.Type == typ }
}

func TestApplyRefusesAPlanOtherThanTheOneSeen(t *testing.T) {
	svc, bundle := planFixture(t)
	req := applyReq(t, svc, bundle)
	req.PlanHash = "another"

	_, err := svc.applyBundle(context.Background(), nil, req, bundle)

	assert.ErrorIs(t, err, hperrors.ErrSpecImportPlanChanged)
}

func TestApplyRefusesABlockedPlan(t *testing.T) {
	svc, bundle := planFixture(t)
	meta, _ := backendRouting(bundle)[specmodel.SettingMetaKey].(map[string]any)
	meta["version"] = 999
	req := applyReq(t, svc, bundle)

	_, err := svc.applyBundle(context.Background(), nil, req, bundle)

	assert.ErrorIs(t, err, hperrors.ErrSpecImportBlocked)
}

func TestApplyRefusesIssuesNobodyAccepted(t *testing.T) {
	svc, bundle := planFixture(t)
	replaceCert(bundle, "fresh", "fresh")
	req := applyReq(t, svc, bundle, backendPath, "-global")
	req.AcceptIssues = false

	_, err := svc.applyBundle(context.Background(), nil, req, bundle)

	assert.ErrorIs(t, err, hperrors.ErrSpecImportIssuesNotAccepted)
}

// Applying the installation's own bundle writes nothing.
func TestApplyOfAnUnchangedInstallationWritesNothing(t *testing.T) {
	svc, bundle := planFixture(t)

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))

	assert.Empty(t, persisted(svc).UpsertingSettings)
	assert.Empty(t, persisted(svc).UpsertingProjects)
	for _, node := range resp.Plan.Nodes {
		if node.Kind != specmodel.NodeKindApp {
			assert.Equal(t, specmodel.OutcomeUnchanged, node.Outcome, node.Path)
		}
	}
}

// A certificate changed in the bundle is the same setting, rewritten; its files
// are written once the transaction commits.
func TestApplyRewritesAChangedCertificateInPlace(t *testing.T) {
	svc, bundle := planFixture(t)
	globalCerts(bundle)["localhost"].(map[string]any)["domain"] = "renewed.example.com"

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))

	cert := persistedSetting(t, svc, ofType(base.SettingTypeSSLCert))
	assert.Equal(t, "cert_1", cert.ID)
	assert.Equal(t, base.ObjectScopeGlobal, cert.Scope)
	assert.Equal(t, 1, cert.UpdateVer)
	assert.Equal(t, "renewed.example.com", cert.MustAsSSLCert().Domain)
	assert.Equal(t, specmodel.OutcomeApplied, node(t, resp.Plan, "global").Outcome)

	assert.NoError(t, resp.AfterCommit(context.Background(), nil))
	assert.Equal(t, []string{"cert_1"}, svc.sslService.(*fakeSSLService).written)
}

// Settings of one import name each other by the ids this import gives them.
func TestApplyGivesACreatedSettingAnIDOthersCanName(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Global.Settings["scripts"] = map[string]any{"build": map[string]any{"data": "make"}}
	bundle.Projects["project_a"].Settings["commandTemplates"] = map[string]any{
		"deploy": map[string]any{"command": "run", "script": map[string]any{"id": "global/scripts/build"}},
	}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	script := persistedSetting(t, svc, ofType(base.SettingTypeScript))
	template := persistedSetting(t, svc, ofType(base.SettingTypeCommandTemplate))
	assert.NotEmpty(t, script.ID)
	assert.Equal(t, "p1", template.ObjectID)
	assert.Equal(t, base.ObjectScopeProject, template.Scope)
	data, err := template.AsCommandTemplate()
	if assert.NoError(t, err) {
		assert.Equal(t, script.ID, data.Script.ID)
	}
}

// A reference validate cleared is written empty, and its setting waits for it.
func TestApplyWritesAClearedReferenceEmptyAndPending(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Projects["project_a"].Settings["commandTemplates"] = map[string]any{
		"deploy": map[string]any{"command": "run", "script": map[string]any{"id": "global/scripts/missing"}},
	}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	template := persistedSetting(t, svc, ofType(base.SettingTypeCommandTemplate))
	assert.Equal(t, base.SettingStatusPending, template.Status)
	data, err := template.AsCommandTemplate()
	if assert.NoError(t, err) {
		assert.Empty(t, data.Script.ID)
	}
}

// A volume pinned to a node this installation lacks is written unpinned, and
// stays active.
func TestApplyUnpinsAVolumeFromAMissingNode(t *testing.T) {
	svc, bundle := planFixture(t)
	volumes, _ := bundle.Projects["project_a"].Settings["volumes"].(map[string]any)
	volumes["default"].(map[string]any)["nodeId"] = "node_gone"

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	volume := persistedSetting(t, svc, ofType(base.SettingTypeClusterVolume))
	assert.Equal(t, "vol_setting_1", volume.ID)
	assert.Equal(t, base.SettingStatusActive, volume.Status)
	assert.Empty(t, volume.MustAsClusterVolume().NodeID)
}

func TestApplyRenamesAnExistingProject(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Projects["project_a"].Name = "Project A, renamed"

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	if projects := persisted(svc).UpsertingProjects; assert.Len(t, projects, 1) {
		assert.Equal(t, "p1", projects[0].ID)
		assert.Equal(t, "Project A, renamed", projects[0].Name)
		assert.Equal(t, 1, projects[0].UpdateVer)
	}
}

func TestApplyAddsAnEnvToAnExistingProject(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Envs["project_a"]["stg"] = &specmodel.EnvDoc{Project: "project_a", Env: "stg", Name: "staging", Index: 1}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	if envs := persisted(svc).UpsertingProjectEnvs; assert.Len(t, envs, 1) {
		assert.Equal(t, []any{"p1:stg", "stg", "staging", 1}, []any{envs[0].ID, envs[0].Key, envs[0].Name, envs[0].Index})
	}
}

// A project from another installation is created as the dashboard creates one,
// and its settings are written over the defaults: a webhook whose secret the
// bundle omitted keeps the one it was created with.
func TestApplyCreatesAProjectFromAnotherInstallation(t *testing.T) {
	svc, bundle := planFixture(t)
	other, err := readBundle(exportedBytes(t, specmodel.SecretsModeOmit, ""), "")
	assert.NoError(t, err)
	project := other.Projects["project_a"]
	project.ID, project.Name = "", "Project New"
	project.Owner = &specmodel.ProjectOwner{ID: "u_elsewhere", Email: "owner@example.com"}
	project.Settings["repoWebhooks"] = map[string]any{"default": map[string]any{
		"kind": "github", "secret": "", specmodel.SettingMetaKey: map[string]any{"name": "default"},
	}}
	for _, app := range other.Envs["project_a"]["dev"].Apps {
		app.ID = ""
	}
	bundle.Projects["project_new"] = project
	bundle.Envs["project_new"] = other.Envs["project_a"]

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle,
		"projects/project_new/settings", "projects/project_new/envs/dev/settings"))

	written := persisted(svc)
	if !assert.Len(t, written.UpsertingProjects, 1) {
		return
	}
	created := written.UpsertingProjects[0]
	assert.Equal(t, []any{"project_new", "Project New", "u1"}, []any{created.Key, created.Name, created.OwnerID})
	if assert.Len(t, written.UpsertingProjectEnvs, 1) {
		assert.Equal(t, created.ID+":dev", written.UpsertingProjectEnvs[0].ID)
	}
	volume := persistedSetting(t, svc, ofType(base.SettingTypeClusterVolume))
	assert.Equal(t, created.ID, volume.ObjectID)
	assert.NotEqual(t, "vol_setting_1", volume.ID, "another installation's id is not reused")

	webhook := persistedSetting(t, svc, ofType(base.SettingTypeRepoWebhook))
	assert.Equal(t, "webhook_"+created.ID, webhook.ID, "the default is written over, not duplicated")
	data, err := webhook.AsRepoWebhook()
	if assert.NoError(t, err) {
		assert.Equal(t, base.WebhookKind("github"), data.Kind)
		plain, _ := data.Secret.GetPlain()
		assert.Equal(t, "generated", plain)
	}
	assert.Equal(t, specmodel.OutcomeApplied, node(t, resp.Plan, "projects/project_new").Outcome)
}
