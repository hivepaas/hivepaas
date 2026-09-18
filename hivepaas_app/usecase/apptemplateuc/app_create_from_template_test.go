package apptemplateuc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

const testTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: pg
  title: PG
  tagline: Test database
  description: Test.
  categories: [databases/sql]
  icon: icons/pg.svg
  requires: {versionCode: v000001}
parameters:
  - {name: username, title: Username, type: string, default: app}
  - {name: password, title: Password, type: secret, generate: {length: 24}}
  - {name: dataVolume, title: Data volume, type: volume}
variants:
  - {name: alpine, title: Alpine, default: true}
versions:
  - {name: "18", release: "18.6", default: true, images: {alpine: "postgres:18.6-alpine3.24"}}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
  settings:
    kind:
      category: database
      engine: postgres
      database: {username: "${{ params.username }}", password: "${{ params.password }}"}
    routing: {port: 5432}
`

var (
	errTestRouting   = errors.New("traefik unavailable")
	errTestProvision = errors.New("provisioning failed")
)

type fakeTemplateService struct {
	apptemplateservice.Service
	resp *apptemplateservice.RenderResp
	err  error
}

func (f *fakeTemplateService) Render(
	context.Context, *apptemplateservice.RenderReq,
) (*apptemplateservice.RenderResp, error) {
	return f.resp, f.err
}

// fakeProvisionService runs Configure the way ProvisionApp does and hands back
// the app with what it returned. Each app it provisions is its own, and one
// named failName fails.
type fakeProvisionService struct {
	appprovisionservice.Service
	called   bool
	names    []string
	failName string
}

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

type fakeSpecService struct {
	specservice.Service
}

func (f *fakeSpecService) BuildApp(
	_ context.Context, _ database.IDB, req *specservice.BuildAppReq,
) (*specservice.BuildAppResp, error) {
	deployment := &entity.Setting{ID: "set-deploy", Type: base.SettingTypeAppDeployment, ObjectID: req.App.ID}
	deployment.MustSetData(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodImage,
		ImageSource:  &entity.DeploymentImageSource{Image: "postgres:18.6-alpine3.24"},
	})
	routing := &entity.Setting{ID: "set-routing", Type: base.SettingTypeAppRouting, ObjectID: req.App.ID}
	routing.MustSetData(&entity.AppRoutingSettings{Port: 5432})
	return &specservice.BuildAppResp{Settings: []*entity.Setting{deployment, routing}}, nil
}

type fakeEnvVarService struct {
	envvarservice.Service
	applied bool
}

func (f *fakeEnvVarService) BuildEnvVarsForAllAppsInScope(
	context.Context, database.IDB, *entity.ObjectScope, bool, []string, bool, bool,
) ([]*envvarservice.AppEnvVarData, error) {
	return nil, nil
}

func (f *fakeEnvVarService) ApplyEnvVarsForApps(
	context.Context, database.IDB, []*envvarservice.AppEnvVarData, bool, bool,
) map[int]error {
	f.applied = true
	return nil
}

type fakeRoutingService struct {
	approutingservice.Service
	req *approutingservice.ApplyAppRoutingReq
	err error
	// failTimes is how many of the first calls fail with err before it succeeds;
	// err alone fails every call.
	failTimes int
	calls     int
}

func (f *fakeRoutingService) ApplyRoutingSettings(
	_ context.Context, _ database.IDB, req *approutingservice.ApplyAppRoutingReq,
) (*approutingservice.ApplyAppRoutingResp, error) {
	f.req = req
	f.calls++
	if f.err != nil && (f.failTimes == 0 || f.calls <= f.failTimes) {
		return nil, f.err
	}
	return &approutingservice.ApplyAppRoutingResp{}, nil
}

type fakeDeploymentService struct {
	appdeploymentservice.Service
}

func (f *fakeDeploymentService) CreateDeploymentAndTask(
	app *entity.App, settings *entity.AppDeploymentSettings,
) (*entity.Deployment, *entity.Task, error) {
	return &entity.Deployment{ID: "dep-1", AppID: app.ID, Settings: settings}, &entity.Task{ID: "task-1"}, nil
}

type fakeAppService struct {
	appservice.Service
	persisted *appservice.PersistingAppData
}

func (f *fakeAppService) PersistAppData(_ context.Context, _ database.IDB, data *appservice.PersistingAppData) error {
	f.persisted = data
	return nil
}

type fakeAuditService struct {
	auditservice.Service
	entries []*auditservice.Entry
}

func (f *fakeAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type createFakes struct {
	templates *fakeTemplateService
	provision *fakeProvisionService
	envVars   *fakeEnvVarService
	routing   *fakeRoutingService
	apps      *fakeAppService
	audit     *fakeAuditService
}

func newCreateTest(t *testing.T) (*UC, *createFakes) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)

	tmpl, err := templatemodel.DecodeTemplate([]byte(testTemplateYAML))
	assert.NoError(t, err)
	result, err := templaterender.Render(&templaterender.Request{
		Template: tmpl,
		Params:   map[string]any{"dataVolume": "vol-1"},
	})
	assert.NoError(t, err)

	fakes := &createFakes{
		templates: &fakeTemplateService{resp: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{
				Source:   "official",
				Revision: "0123456789abcdef0123456789abcdef01234567",
				Entry:    &templatemodel.IndexEntry{Name: "pg", File: templatemodel.FileRef{SHA256: "file-sha"}},
				Template: tmpl,
			},
			Result: result,
		}},
		provision: &fakeProvisionService{},
		envVars:   &fakeEnvVarService{},
		routing:   &fakeRoutingService{},
		apps:      &fakeAppService{},
		audit:     &fakeAuditService{},
	}
	uc := &UC{
		appDeploymentService: &fakeDeploymentService{},
		appProvisionService:  fakes.provision,
		appRoutingService:    fakes.routing,
		appService:           fakes.apps,
		appTemplateService:   fakes.templates,
		auditService:         fakes.audit,
		envVarService:        fakes.envVars,
		specService:          &fakeSpecService{},
	}
	return uc, fakes
}

func testCreateReq() *apptemplatedto.CreateAppFromTemplateReq {
	return &apptemplatedto.CreateAppFromTemplateReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "main-db", Template: "pg",
		Params: map[string]any{"dataVolume": "vol-1"},
	}
}

func testAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "user-1"}}}
}

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

func errDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestProvisionFromTemplateRecordsTheTemplate(t *testing.T) {
	uc, fakes := newCreateTest(t)

	created := provision(t, uc, fakes)

	var binding *entity.Setting
	for _, setting := range created.app.Settings {
		if setting.Type == base.SettingTypeAppTemplate {
			binding = setting
		}
	}
	if binding == nil {
		t.Fatal("the app records the template it was provisioned from")
	}
	assert.Equal(t, base.ObjectScopeApp, binding.Scope)
	assert.Equal(t, "app-1", binding.ObjectID)

	stored := &entity.Setting{Type: base.SettingTypeAppTemplate, Data: binding.Data}
	data, err := stored.AsAppTemplateSettings()
	assert.NoError(t, err)
	assert.Equal(t, "official", data.Source)
	assert.Equal(t, "pg", data.Template)
	assert.Equal(t, "PG", data.Title)
	assert.Equal(t, "18", data.Version)
	assert.Equal(t, "alpine", data.Variant)
	assert.Equal(t, "18.6", data.Base.Release)
	assert.Equal(t, "file-sha", data.Base.TemplateSHA256)
	assert.Equal(t, fakes.templates.resp.Result.BaseSHA256, data.Base.RenderedSHA256)
	assert.Equal(t, "app", data.Params["username"].Value)
	assert.Equal(t, "vol-1", data.Params["dataVolume"].Value)

	generated := fakes.templates.resp.Result.Params["password"].Text()
	assert.NotContains(t, binding.Data, generated, "a secret parameter is stored encrypted")
	assert.NotContains(t, data.Base.Rendered, generated, "the base holds no secret")
	password, err := data.Params["password"].Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, generated, password)
}

func TestProvisionFromTemplateAppliesAndDeploys(t *testing.T) {
	uc, fakes := newCreateTest(t)

	created := provision(t, uc, fakes)

	assert.True(t, fakes.envVars.applied)
	assert.Equal(t, 5432, fakes.routing.req.RoutingSettings.Port)
	assert.Equal(t, []*entity.Deployment{created.deployment}, fakes.apps.persisted.UpsertingDeployments)
	assert.Equal(t, []*entity.Task{created.deploymentTask}, fakes.apps.persisted.UpsertingTasks)
	assert.Equal(t, "postgres:18.6-alpine3.24", created.deployment.Settings.ImageSource.Image)
	assert.Equal(t, base.DeploymentTriggerSourceAPI, created.deployment.Trigger.Source)
	assert.Equal(t, "user-1", created.deployment.Trigger.SourceID)
}

func TestProvisionFromTemplateAuditsWithoutSecrets(t *testing.T) {
	uc, fakes := newCreateTest(t)

	provision(t, uc, fakes)

	assert.Len(t, fakes.audit.entries, 1)
	entry := fakes.audit.entries[0]
	assert.Equal(t, base.AuditLogTypeAppCreate, entry.Type)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Contains(t, entry.Detail, `"template":"pg"`)
	assert.Contains(t, entry.Detail, `"revision":"0123456789abcdef0123456789abcdef01234567"`)
	assert.NotContains(t, entry.Detail, fakes.templates.resp.Result.Params["password"].Text())
}

func TestProvisionFromTemplateReturnsWhatItCreatedWhenItFails(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.routing.err = errTestRouting

	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		singleApp(fakes))

	assert.ErrorIs(t, err, errTestRouting)
	assert.Equal(t, "svc-main-db", created.app.ServiceID, "the caller needs the service to remove it")
}

// Swarm is still writing to a service it has just created, so the first routing
// update can be refused as out of sequence. Failing the create for that would
// leave a user with a template that works two times in three.
func TestProvisionFromTemplateRetriesRoutingWhileSwarmSettles(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.routing.err = errTestRouting
	fakes.routing.failTimes = routingApplyRetryMax

	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		singleApp(fakes))

	assert.NoError(t, err)
	assert.Equal(t, routingApplyRetryMax+1, fakes.routing.calls)
	assert.NotNil(t, created.deployment, "the create carries on once routing is applied")
}

func TestCreateAppFromTemplateRefusesBeforeProvisioning(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.templates.err = hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid)

	_, err := uc.CreateAppFromTemplate(context.Background(), testAuth(), testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
	assert.False(t, fakes.provision.called)
}

func TestCreateAppFromTemplateRefusesATemplateThatDeploysNothing(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.templates.resp.Result.Doc.Deployment = nil

	_, err := uc.CreateAppFromTemplate(context.Background(), testAuth(), testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	assert.False(t, fakes.provision.called)
}

func TestTransformAppTemplateBinding(t *testing.T) {
	appliedAt := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	resp := apptemplatedto.TransformAppTemplateBinding(&entity.AppTemplateSettings{
		Source: "official", Template: "pg", Title: "PG", Version: "18", Variant: "alpine",
		ImageOverride: "postgres:18.7-alpine3.24",
		Params:        map[string]*entity.AppTemplateParam{"password": {Secret: entity.NewEncryptedField("x")}},
		Base:          entity.AppTemplateBase{Revision: "abc", Release: "18.6", AppliedAt: appliedAt},
	})
	assert.Equal(t, &apptemplatedto.AppTemplateBindingResp{
		Source: "official", Template: "pg", Title: "PG", Version: "18", Release: "18.6",
		Variant: "alpine", Revision: "abc", AppliedAt: appliedAt,
		ImageOverride: "postgres:18.7-alpine3.24",
		Dependencies:  []*apptemplatedto.AppTemplateBindingDependencyResp{},
	}, resp, "parameters are not part of the response")
}

func TestProvisionFromTemplateRecordsAnImageOverride(t *testing.T) {
	uc, fakes := newCreateTest(t)
	rendered := fakes.templates.resp
	rendered.Result.ImageOverride = "postgres:18.7-alpine3.24"
	rendered.Result.ImageOverrideClass = templatemodel.ImageOverrideSameLine

	created := provision(t, uc, fakes)

	var binding *entity.Setting
	for _, setting := range created.app.Settings {
		if setting.Type == base.SettingTypeAppTemplate {
			binding = setting
		}
	}
	stored := &entity.Setting{Type: base.SettingTypeAppTemplate, Data: binding.Data}
	data, err := stored.AsAppTemplateSettings()
	assert.NoError(t, err)
	assert.Equal(t, "postgres:18.7-alpine3.24", data.ImageOverride)
	assert.Equal(t, "18", data.Version, "the version still describes the template, not the override")

	assert.Contains(t, fakes.audit.entries[0].Detail, `"imageOverride":"postgres:18.7-alpine3.24"`)
}

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

// fakeClusterSecretService fills in what docker would have returned: an id for
// every entry that asked to be mounted as a file, and nothing for the rest.
type fakeClusterSecretService struct {
	clustersecretservice.Service
	secrets []*entity.Secret
	configs []*entity.ConfigFile
}

func (f *fakeClusterSecretService) CreateSecretsForApp(
	_ context.Context, _ database.IDB, _ *entity.App, secrets []*entity.Secret,
) ([]*entity.SwarmSecretRef, error) {
	f.secrets = secrets
	refs := make([]*entity.SwarmSecretRef, 0, len(secrets))
	for _, secret := range secrets {
		if secret.SwarmRef == nil || secret.SwarmRef.File == nil {
			refs = append(refs, nil)
			continue
		}
		secret.SwarmRef.SecretID = "docker-secret-" + secret.Key
		refs = append(refs, secret.SwarmRef)
	}
	return refs, nil
}

func (f *fakeClusterSecretService) CreateConfigsForApp(
	_ context.Context, _ database.IDB, _ *entity.App, configs []*entity.ConfigFile,
) ([]*entity.SwarmConfigRef, error) {
	f.configs = configs
	refs := make([]*entity.SwarmConfigRef, 0, len(configs))
	for _, config := range configs {
		if config.SwarmRef == nil || config.SwarmRef.File == nil {
			refs = append(refs, nil)
			continue
		}
		config.SwarmRef.ConfigID = "docker-config-" + config.Name
		refs = append(refs, config.SwarmRef)
	}
	return refs, nil
}

func secretSetting(t *testing.T, id string, secret *entity.Secret) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: base.SettingTypeSecret, Name: secret.Key, ObjectID: "app-1"}
	assert.NoError(t, setting.SetData(secret))
	return setting
}

func configFileSetting(t *testing.T, id string, configFile *entity.ConfigFile) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: base.SettingTypeConfigFile, Name: configFile.Name, ObjectID: "app-1"}
	assert.NoError(t, setting.SetData(configFile))
	return setting
}

func TestApplySecretsAndConfigFilesKeepsTheIDsDockerCreatedThemWith(t *testing.T) {
	uc, fakes := newCreateTest(t)
	cluster := &fakeClusterSecretService{}
	uc.clusterSecretService = cluster
	app := &entity.App{ID: "app-1", ServiceID: "svc-1", Settings: []*entity.Setting{
		secretSetting(t, "set-env", &entity.Secret{Key: "ADMIN_PASSWORD", Value: entity.NewEncryptedField("s3cret")}),
		secretSetting(t, "set-file", &entity.Secret{Key: "LICENSE_KEY", Value: entity.NewEncryptedField("abc"),
			SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "/etc/app/license"}}}),
		configFileSetting(t, "set-conf", &entity.ConfigFile{Name: "app.conf", Content: "listen = 8080",
			SwarmRef: &entity.SwarmConfigRef{File: &entity.SwarmRefFileTarget{Name: "/etc/app/app.conf"}}}),
	}}

	assert.NoError(t, uc.applySecretsAndConfigFiles(t.Context(), nil, app))

	// Both secrets are handed over: which of them becomes a docker object is the
	// cluster service's decision, taken from the file target.
	assert.Len(t, cluster.secrets, 2)
	assert.Len(t, cluster.configs, 1)

	// What docker returned is written back to the settings, so the app can find
	// its objects again.
	assert.Len(t, fakes.apps.persisted.UpsertingSettings, 3)
	stored := settinghelper.FindSettingByType(fakes.apps.persisted.UpsertingSettings, base.SettingTypeSecret)
	assert.Equal(t, "ADMIN_PASSWORD", stored.Name)
	envSecret, err := stored.AsSecret()
	assert.NoError(t, err)
	assert.Nil(t, envSecret.SwarmRef)

	fileSecret, err := fakes.apps.persisted.UpsertingSettings[1].AsSecret()
	assert.NoError(t, err)
	assert.Equal(t, "docker-secret-LICENSE_KEY", fileSecret.SwarmRef.SecretID)
	configFile, err := fakes.apps.persisted.UpsertingSettings[2].AsConfigFile()
	assert.NoError(t, err)
	assert.Equal(t, "docker-config-app.conf", configFile.SwarmRef.ConfigID)

	// The value survives the round trip through the setting, still encrypted.
	assert.NotContains(t, fakes.apps.persisted.UpsertingSettings[1].Data, "abc")
	plain, err := fileSecret.Value.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "abc", plain)
}

func TestApplySecretsAndConfigFilesDoesNothingWithoutThem(t *testing.T) {
	uc, fakes := newCreateTest(t)
	cluster := &fakeClusterSecretService{}
	uc.clusterSecretService = cluster
	app := &entity.App{ID: "app-1", ServiceID: "svc-1", Settings: []*entity.Setting{
		{ID: "set-routing", Type: base.SettingTypeAppRouting, ObjectID: "app-1"},
	}}

	assert.NoError(t, uc.applySecretsAndConfigFiles(t.Context(), nil, app))

	assert.Nil(t, cluster.secrets)
	assert.Nil(t, cluster.configs)
	assert.Nil(t, fakes.apps.persisted)
}
