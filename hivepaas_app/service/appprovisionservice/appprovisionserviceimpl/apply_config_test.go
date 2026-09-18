package appprovisionserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeTaskQueue struct {
	queue.TaskQueue
	scheduled []*entity.Setting
}

func (f *fakeTaskQueue) ScheduleTasksForSchedJobs(
	_ context.Context, _ database.Tx, schedJobs []*entity.Setting, _ bool,
) error {
	f.scheduled = append(f.scheduled, schedJobs...)
	return nil
}

func settingWithData(t *testing.T, id string, data entity.SettingData) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: data.GetType(), ObjectID: "app-1"}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func appWithSettings(settings ...*entity.Setting) *entity.App {
	return &entity.App{ID: "app-1", Name: "db", ServiceID: "svc-1", Settings: settings}
}

func TestApplyAppConfigurationKeepsTheIDsDockerCreatedThemWith(t *testing.T) {
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	svc, fakes := newProvisionTest(t)
	envOnly := settingWithData(t, "set-env",
		&entity.Secret{Key: "ADMIN_PASSWORD", Value: entity.NewEncryptedField("s3cret")})
	onFile := settingWithData(t, "set-file", &entity.Secret{Key: "LICENSE_KEY",
		Value:    entity.NewEncryptedField("abc"),
		SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "/etc/app/license"}}})
	configFile := settingWithData(t, "set-conf", &entity.ConfigFile{Name: "app.conf", Content: "listen = 8080",
		SwarmRef: &entity.SwarmConfigRef{File: &entity.SwarmRefFileTarget{Name: "/etc/app/app.conf"}}})
	app := appWithSettings(envOnly, onFile, configFile)

	resp, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.True(t, fakes.envVars.applied)
	// Both secrets are handed over: which of them becomes a docker object is the
	// cluster service's decision, taken from the file target.
	assert.Len(t, fakes.clusterFiles.secrets, 2)
	assert.Len(t, fakes.clusterFiles.configs, 1)
	assert.Equal(t, []*entity.SwarmConfigRef{configFile.MustAsConfigFile().SwarmRef}, resp.Configs)
	assert.Nil(t, resp.Secrets[0], "a secret read through the environment makes nothing in docker")

	// What docker returned is written back to the settings, so the app can find
	// its objects again. Reading it from Data is the point: the parsed object
	// carries it either way, and only Data is what the database keeps.
	assert.Len(t, fakes.apps.persisted.UpsertingSettings, 3)
	stored := func(setting *entity.Setting) *entity.Setting {
		return &entity.Setting{Type: setting.Type, Data: setting.Data}
	}
	assert.Equal(t, "docker-config-app.conf", stored(configFile).MustAsConfigFile().SwarmRef.ConfigID)
	assert.Equal(t, "docker-secret-LICENSE_KEY", stored(onFile).MustAsSecret().SwarmRef.SecretID)
	assert.Nil(t, stored(envOnly).MustAsSecret().SwarmRef)

	// The value survives the round trip through the setting, still encrypted.
	assert.NotContains(t, onFile.Data, "abc")
	plain, err := stored(onFile).MustAsSecret().Value.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "abc", plain)
}

func TestApplyAppConfigurationTouchesDockerOnlyForWhatTheAppHas(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	app := appWithSettings(&entity.Setting{ID: "set-kind", Type: base.SettingTypeAppKind, ObjectID: "app-1"})

	resp, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.Empty(t, resp.Secrets)
	assert.Empty(t, resp.Configs)
	assert.Nil(t, fakes.clusterFiles.secrets)
	assert.Nil(t, fakes.clusterFiles.configs)
	assert.Nil(t, fakes.apps.persisted, "nothing to write back")
	assert.Zero(t, fakes.routing.calls, "an app with no routing settings is not routed")
}

// Swarm is still writing to a service it has just created, so the first routing
// update can be refused as out of sequence. Failing the create for that would
// leave a user with a template that works two times in three.
func TestApplyAppConfigurationRetriesRoutingWhileSwarmSettles(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	fakes.routing.err = errTestRouting
	fakes.routing.failTimes = routingApplyRetryMax
	app := appWithSettings(settingWithData(t, "set-routing", &entity.AppRoutingSettings{Port: 5432}))

	_, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.Equal(t, routingApplyRetryMax+1, fakes.routing.calls)
	assert.Equal(t, 5432, fakes.routing.req.RoutingSettings.Port)
}

func TestApplyAppConfigurationGivesUpOnRoutingThatKeepsFailing(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	fakes.routing.err = errTestRouting
	app := appWithSettings(settingWithData(t, "set-routing", &entity.AppRoutingSettings{Port: 5432}))

	_, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.ErrorIs(t, err, errTestRouting)
	assert.Equal(t, routingApplyRetryMax+1, fakes.routing.calls)
}

func TestApplyAppConfigurationSchedulesTheAppsJobs(t *testing.T) {
	svc, _ := newProvisionTest(t)
	taskQueue := &fakeTaskQueue{}
	svc.taskQueue = taskQueue
	job := &entity.Setting{ID: "set-job", Type: base.SettingTypeSchedJob, ObjectID: "app-1"}

	_, err := svc.ApplyAppConfiguration(context.Background(), database.Tx{},
		&appprovisionservice.ApplyAppConfigurationReq{App: appWithSettings(job)})

	assert.NoError(t, err)
	assert.Equal(t, []*entity.Setting{job}, taskQueue.scheduled)
}

func routedApp(t *testing.T, domain string) *entity.App {
	t.Helper()
	return appWithSettings(settingWithData(t, "set-routing", &entity.AppRoutingSettings{
		Port: 2368, ExposePublicly: true,
		Domains: []*entity.AppDomain{{Domain: domain, Enabled: true, Protocol: base.NetworkProtocolHTTP}},
	}))
}

func certSetting(id, domain string) *entity.Setting {
	setting := &entity.Setting{ID: id, Type: base.SettingTypeSSLCert, Name: domain}
	setting.MustSetData(&entity.SSLCert{Domain: domain})
	return setting
}

// An app created with a domain has nobody to pick a certificate for it, and a
// wildcard is usually why the domain could be handed out at all.
func TestApplyAppConfigurationAttachesACertificateTheSystemAlreadyHas(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	cert := certSetting("ssl-1", "*.example.com")
	fakes.domains.certs = map[string]*entity.Setting{"blog.example.com": cert}
	app := routedApp(t, "blog.example.com")

	_, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.Equal(t, []string{"blog.example.com"}, fakes.domains.asked)
	assert.Equal(t, "ssl-1", fakes.routing.req.RoutingSettings.Domains[0].SSLCert.ID)
	assert.Same(t, cert, fakes.routing.req.RefObjects.RefSettings["ssl-1"],
		"the certificate has to be loaded for the routing service to write its files")

	// The choice is the app's from now on, so it is written back rather than
	// re-decided on every deployment.
	stored := &entity.Setting{Type: base.SettingTypeAppRouting,
		Data: fakes.apps.persisted.UpsertingSettings[0].Data}
	assert.Equal(t, "ssl-1", stored.MustAsAppRoutingSettings().Domains[0].SSLCert.ID)
}

func TestApplyAppConfigurationRoutesADomainNothingCovers(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	app := routedApp(t, "blog.example.com")

	_, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.Empty(t, fakes.routing.req.RoutingSettings.Domains[0].SSLCert.ID)
	assert.Nil(t, fakes.apps.persisted, "nothing was chosen, so nothing is written back")
	assert.Equal(t, 1, fakes.routing.calls, "the app is still routed")
}

func TestApplyAppConfigurationLooksForNoCertificateWithoutADomain(t *testing.T) {
	svc, fakes := newProvisionTest(t)
	app := appWithSettings(settingWithData(t, "set-routing", &entity.AppRoutingSettings{Port: 2368}))

	_, err := svc.ApplyAppConfiguration(context.Background(), nil,
		&appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.Empty(t, fakes.domains.asked)
}
