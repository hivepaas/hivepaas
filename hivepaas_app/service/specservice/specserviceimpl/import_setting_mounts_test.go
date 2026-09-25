package specserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const mountChange = "settings.settingMounts/cert"

func backendMount(bundle *specmodel.ImportBundle) map[string]any {
	entries, _ := backendSettings(bundle)["settingMounts"].(map[string]any)
	entry, _ := entries["cert"].(map[string]any)
	return entry
}

func addPrivateKey(bundle *specmodel.ImportBundle) {
	entry := backendMount(bundle)
	files, _ := entry["files"].([]any)
	entry["files"] = append(files, map[string]any{"part": "privateKey", "path": "/etc/app/tls/key.pem"})
}

func TestExportWritesAnEntryWithItsSourceAndFiles(t *testing.T) {
	_, bundle := planFixture(t)

	entry := backendMount(bundle)

	if assert.NotNil(t, entry) {
		assert.NotNil(t, entry["source"])
		assert.Len(t, entry["files"], 1)
	}
}

func planAskingMounts(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, allowed bool,
) (*specmodel.ImportPlan, *int) {
	t.Helper()
	asked := 0
	out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
		MayMountSecrets: func(context.Context) (bool, error) { asked++; return allowed, nil },
	})
	return out, &asked
}

// The rest of the app is imported; the entry that hands out a private key is
// not, for a caller who may not reveal secrets.
func TestPlanSkipsAMountOfAPrivateKeyForAnOperatorWhoMayNotRevealSecrets(t *testing.T) {
	for allowed, wantChange := range map[bool]bool{false: false, true: true} {
		svc, bundle := planFixture(t)
		addPrivateKey(bundle)

		out, asked := planAskingMounts(t, svc, bundle, allowed)

		backend := node(t, out, backendPath)
		assert.Equal(t, 1, *asked)
		assert.Equal(t, wantChange, contains(backend.Changes, mountChange), "allowed=%v", allowed)
		if !wantChange {
			issues := issuesOf(backend, specmodel.CodeSettingMountNotPermitted)
			if assert.Len(t, issues, 1) {
				assert.Equal(t, specmodel.SeveritySkipped, issues[0].Severity)
			}
		}
	}
}

// The installed entry and the bundle's agree: nothing new is handed out.
func TestPlanDoesNotAskForAnEntryThatGrantsNothingNew(t *testing.T) {
	svc, bundle := planFixture(t)
	backendMount(bundle)["files"] = []any{map[string]any{"part": "certificate", "path": "/srv/cert.pem"}}

	out, asked := planAskingMounts(t, svc, bundle, false)

	assert.Zero(t, *asked, "a path moved; no pair was added")
	assert.Contains(t, node(t, out, backendPath).Changes, mountChange)
}

func TestApplyLeavesARefusedEntryOut(t *testing.T) {
	svc, bundle := planFixture(t)
	addPrivateKey(bundle)
	req := applyReq(t, svc, bundle)
	req.MayMountSecrets = func(context.Context) (bool, error) { return false, nil }
	req.AcceptIssues = true
	req.PlanHash = planWith(t, svc, bundle, &req.ValidateImportReq).PlanHash

	apply(t, svc, bundle, req)

	for _, setting := range persisted(svc).UpsertingSettings {
		assert.NotEqual(t, base.SettingTypeAppSettingMount, setting.Type, "the refused entry is not written")
	}
}

// An app created by the import writes every setting it holds; a refused entry
// is still left out.
func TestApplyLeavesARefusedEntryOutOfANewApp(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, map[string]any{"settingMounts": map[string]any{"key": map[string]any{
		"source": backendMount(bundle)["source"],
		"files":  []any{map[string]any{"part": "privateKey", "path": "/etc/key.pem"}},
	}}})
	req := applyReq(t, svc, bundle)
	req.MayMountSecrets = func(context.Context) (bool, error) { return false, nil }
	req.AcceptIssues = true
	req.PlanHash = planWith(t, svc, bundle, &req.ValidateImportReq).PlanHash

	apply(t, svc, bundle, req)

	provision := svc.appProvisionService.(*fakeProvisionService)
	for _, settings := range provision.settings {
		for _, setting := range settings {
			assert.NotEqual(t, base.SettingTypeAppSettingMount, setting.Type)
		}
	}
	_ = entity.CurrentAppSettingMountVersion
}
