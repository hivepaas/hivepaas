package specserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func plainOf(t *testing.T, field entity.EncryptedField) string {
	t.Helper()
	plain, err := field.GetPlain()
	assert.NoError(t, err)
	return plain
}

// A domain another app holds is dropped from the routing written; the rest of
// the routing is.
func TestApplyDropsADomainAnotherAppHolds(t *testing.T) {
	svc, bundle := planFixture(t)
	svc.domainService.(*fakeDomainService).held = map[string]string{"api.example.com": "app_other"}
	backendRouting(bundle)["exposePublicly"] = true
	backendRouting(bundle)["port"] = 9090

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	routing := persistedSetting(t, svc, ofType(base.SettingTypeAppRouting))
	assert.Equal(t, "routing_1", routing.ID)
	data := routing.MustAsAppRoutingSettings()
	assert.Equal(t, 9090, data.Port)
	assert.Empty(t, data.Domains)
}

// A secret written over one of an omit bundle keeps its value. A swarmRef in the
// bundle - an older export - is not carried over: secrets mount through setting
// mounts.
func TestApplyKeepsASecretsValueAndDropsASwarmRef(t *testing.T) {
	svc, bundle := planFixture(t)
	secrets, _ := backendSettings(bundle)["secrets"].(map[string]any)
	secret, _ := secrets["db-password"].(map[string]any)
	secret["swarmRef"] = map[string]any{
		"file": map[string]any{"name": "db"}, "secretId": "elsewhere", "secretName": "elsewhere",
	}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	written := persistedSetting(t, svc, ofType(base.SettingTypeSecret))
	assert.Equal(t, "secret_1", written.ID)
	data, err := written.AsSecret()
	if assert.NoError(t, err) {
		assert.Equal(t, "hunter2", plainOf(t, data.Value))
	}
	assert.NotContains(t, written.Data, "swarmRef")
}

// An app matched by key keeps the credential its data was initialized with.
func TestApplyKeepsTheCredentialOfAnAppMatchedByKey(t *testing.T) {
	svc, bundle := encryptedPlanFixture(t)
	bundle.Envs["project_a"]["dev"].Apps["backend"].ID = "app_elsewhere"
	kind, _ := backendSettings(bundle)["kind"].(map[string]any)
	kind["database"].(map[string]any)["password"] = "another"
	kind["version"] = "17"

	req := applyReq(t, svc, bundle)
	req.Passphrase = bundlePassphrase
	apply(t, svc, bundle, req)

	written := persistedSetting(t, svc, ofType(base.SettingTypeAppKind))
	assert.Equal(t, "kind_1", written.ID)
	data := written.MustAsAppKindSettings()
	assert.Equal(t, "17", data.Version)
	assert.Equal(t, "s3cret", plainOf(t, data.Database.Password))
}

// A webhook created from a bundle without secrets gets one: HivePaaS owns it.
func TestApplyGeneratesTheSecretOfACreatedWebhook(t *testing.T) {
	svc, bundle := planFixture(t)
	bundle.Projects["project_a"].Settings["repoWebhooks"] = map[string]any{
		"ci": map[string]any{"kind": "github", "secret": ""},
	}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	webhook := persistedSetting(t, svc, ofType(base.SettingTypeRepoWebhook))
	data, err := webhook.AsRepoWebhook()
	if assert.NoError(t, err) {
		assert.Len(t, plainOf(t, data.Secret), 2*generatedSecretBytes)
	}
}

// An app created from a bundle without secrets gets its credential generated,
// and its settings are kept for provisioning rather than persisted.
func TestApplyGeneratesTheCredentialOfACreatedApp(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, map[string]any{"kind": map[string]any{
		"category": "database", "engine": "postgres", "database": map[string]any{"dbName": "worker"},
	}})
	req := applyReq(t, svc, bundle, workerPath)
	p, err := svc.planBundle(context.Background(), nil, &req.ValidateImportReq, bundle)
	assert.NoError(t, err)
	w := newWriter(p, "u_operator")

	assert.NoError(t, w.write(context.Background()))

	for _, setting := range svc.projectService.(*fakeProjectService).persisted.UpsertingSettings {
		assert.NotEqual(t, base.SettingTypeAppKind, setting.Type, "not persisted with the scopes")
	}
	settings := w.appSettings[workerPath]
	if assert.Len(t, settings, 1) {
		kind := settings[0].MustAsAppKindSettings()
		assert.Equal(t, w.appIDs[workerPath], settings[0].ObjectID)
		assert.Equal(t, base.ObjectScopeApp, settings[0].Scope)
		assert.Len(t, plainOf(t, kind.Database.Password), 2*generatedSecretBytes)
		assert.NotEmpty(t, plainOf(t, kind.Database.RootPassword))
	}
}

// Everything appKindCredentials names is generated.
func TestGenerateOwnedSecretsCoversEveryCredential(t *testing.T) {
	kind := &entity.AppKindSettings{
		Database: &entity.AppKindDatabase{}, Cache: &entity.AppKindCache{}, Storage: &entity.AppKindStorage{},
	}
	assert.Equal(t, len(appKindCredentials), entity.CountEmptySecrets(kind))

	generateOwnedSecrets(kind)

	assert.Zero(t, entity.CountEmptySecrets(kind))
}

// A mount is built from the volume's id here: the target's for a path, what an
// external reference finds; one nothing satisfies, and a port another service
// holds, are left out.
func TestPreparedDeploymentNamesVolumesByIDAndDropsWhatValidateCleared(t *testing.T) {
	svc, bundle := planFixture(t)
	svc.clusterService.(*fakeClusterService).ports = map[clusterservice.PortRef]string{
		{Published: 8080, Protocol: network.TCP}: "svc_other",
	}
	addWorker(bundle, nil)
	worker := bundle.Envs["project_a"]["dev"].Apps["worker"]
	worker.Deployment.Storage.Mounts["/shared"] = specmodel.Mount{Type: mount.TypeVolume,
		External: &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared", ID: "gvol_1"}}
	worker.Deployment.Storage.Mounts["/lost"] = specmodel.Mount{Type: mount.TypeVolume,
		External: &specmodel.ExternalRef{Type: "cluster-volume", Name: "nowhere"}}
	worker.Deployment.Networks = &specmodel.Networks{EndpointSpec: &specmodel.EndpointSpec{
		Ports: []*specmodel.PortConfig{{Target: 80, Published: 8080}, {Target: 81, Published: 8081}},
	}}
	req := applyReq(t, svc, bundle, workerPath)
	p, err := svc.planBundle(context.Background(), nil, &req.ValidateImportReq, bundle)
	assert.NoError(t, err)
	w := newWriter(p, "u_operator")
	assert.NoError(t, w.write(context.Background()))

	prepared, err := w.preparedDeployment(context.Background(), p.byPath[workerPath])

	assert.NoError(t, err)
	assert.Nil(t, prepared.Source)
	assert.Equal(t, "vol_setting_1", prepared.Storage.Mounts["/data"].Source)
	assert.Equal(t, "gvol_1", prepared.Storage.Mounts["/shared"].Source)
	assert.Nil(t, prepared.Storage.Mounts["/shared"].External)
	assert.NotContains(t, prepared.Storage.Mounts, "/lost")
	if assert.Len(t, prepared.Networks.EndpointSpec.Ports, 1) {
		assert.Equal(t, uint32(8081), prepared.Networks.EndpointSpec.Ports[0].Published)
	}
	assert.Len(t, worker.Deployment.Networks.EndpointSpec.Ports, 2, "the bundle is left as it was")
}
