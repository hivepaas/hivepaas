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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
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

var errTestRouting = errors.New("traefik unavailable")

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
// the app with what it returned.
type fakeProvisionService struct {
	appprovisionservice.Service
	called bool
}

func (f *fakeProvisionService) ProvisionApp(
	ctx context.Context, db database.IDB, req *appprovisionservice.ProvisionAppReq,
) (*appprovisionservice.ProvisionAppResp, error) {
	f.called = true
	app := &entity.App{
		ID: "app-1", Name: req.Name, Key: "main-db", ServiceID: "svc-1",
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

func provision(t *testing.T, uc *UC, fakes *createFakes) *createdFromTemplate {
	t.Helper()
	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		fakes.templates.resp)
	assert.NoError(t, err)
	return created
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
		fakes.templates.resp)

	assert.ErrorIs(t, err, errTestRouting)
	assert.Equal(t, "svc-1", created.app.ServiceID, "the caller needs the service to remove it")
}

// Swarm is still writing to a service it has just created, so the first routing
// update can be refused as out of sequence. Failing the create for that would
// leave a user with a template that works two times in three.
func TestProvisionFromTemplateRetriesRoutingWhileSwarmSettles(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.routing.err = errTestRouting
	fakes.routing.failTimes = routingApplyRetryMax

	created, err := uc.provisionFromTemplate(context.Background(), nil, testAuth(), testCreateReq(),
		fakes.templates.resp)

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
		Params: map[string]*entity.AppTemplateParam{"password": {Secret: entity.NewEncryptedField("x")}},
		Base:   entity.AppTemplateBase{Revision: "abc", Release: "18.6", AppliedAt: appliedAt},
	})
	assert.Equal(t, &apptemplatedto.AppTemplateBindingResp{
		Source: "official", Template: "pg", Title: "PG", Version: "18", Release: "18.6",
		Variant: "alpine", Revision: "abc", AppliedAt: appliedAt,
	}, resp, "parameters are not part of the response")
}
