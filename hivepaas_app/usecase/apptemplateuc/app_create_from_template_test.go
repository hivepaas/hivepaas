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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
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

var errTestProvision = errors.New("provisioning failed")

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

// fakeProvisionService runs each request's Configure the way provisioning does
// and hands back the app with what it returned, so that a test sees the settings
// the use case asked for. One app named failName fails.
type fakeProvisionService struct {
	appprovisionservice.Service
	called    bool
	names     []string
	failName  string
	deployFor []string
}

func (f *fakeProvisionService) ProvisionApps(
	ctx context.Context, db database.IDB, req *appprovisionservice.ProvisionAppsReq,
) (*appprovisionservice.ProvisionAppsResp, error) {
	resp := &appprovisionservice.ProvisionAppsResp{
		Apps:    make([]*appprovisionservice.ProvisionAppResp, 0, len(req.Apps)),
		Cleanup: func(context.Context) error { return nil },
	}
	for _, appReq := range req.Apps {
		one, err := f.ProvisionApp(ctx, db, appReq)
		if one != nil {
			resp.Apps = append(resp.Apps, one)
		}
		if err != nil {
			return resp, hperrors.Wrap(err).WithExtraDetail("while creating %s", appReq.Name)
		}
	}
	return resp, nil
}

func (f *fakeProvisionService) ProvisionApp(
	ctx context.Context, db database.IDB, req *appprovisionservice.ProvisionAppReq,
) (*appprovisionservice.ProvisionAppResp, error) {
	f.called = true
	f.names = append(f.names, req.Name)
	if req.Deployment != nil {
		f.deployFor = append(f.deployFor, req.Name)
	}
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
	resp := &appprovisionservice.ProvisionAppResp{App: app,
		Created: &appprovisionservice.CreatedInDocker{ServiceID: app.ServiceID}}
	if req.Deployment != nil {
		resp.Deployment = &entity.Deployment{ID: "dep-" + app.Key, AppID: app.ID,
			Trigger: &entity.AppDeploymentTrigger{
				Source: req.Deployment.Source, SourceID: req.Deployment.SourceID,
			}}
		resp.DeploymentTask = &entity.Task{ID: "task-" + app.Key}
	}
	return resp, nil
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
		audit:     &fakeAuditService{},
	}
	uc := &UC{
		appProvisionService: fakes.provision,
		appTemplateService:  fakes.templates,
		auditService:        fakes.audit,
		specService:         &fakeSpecService{},
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

// provisionOne creates the request's single app and returns what provisioning
// was handed back for it.
func provisionOne(t *testing.T, uc *UC, fakes *createFakes) *appprovisionservice.ProvisionAppResp {
	t.Helper()
	req := testCreateReq()
	provisioned, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, fakes.templates.resp))
	assert.NoError(t, err)
	assert.Len(t, provisioned.Apps, 1)
	return provisioned.Apps[0]
}

// binding reads back the template an app was provisioned from, from the stored
// form rather than the parsed one.
func binding(t *testing.T, app *entity.App) *entity.AppTemplateSettings {
	t.Helper()
	setting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppTemplate)
	if setting == nil {
		t.Fatal("the app records the template it was provisioned from")
	}
	stored := &entity.Setting{Type: base.SettingTypeAppTemplate, Data: setting.Data}
	data, err := stored.AsAppTemplateSettings()
	assert.NoError(t, err)
	return data
}

func errDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestCreateFromTemplateRecordsTheTemplate(t *testing.T) {
	uc, fakes := newCreateTest(t)

	app := provisionOne(t, uc, fakes).App

	setting := settinghelper.FindSettingByType(app.Settings, base.SettingTypeAppTemplate)
	assert.Equal(t, base.ObjectScopeApp, setting.Scope)
	assert.Equal(t, app.ID, setting.ObjectID)

	data := binding(t, app)
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
	assert.NotContains(t, setting.Data, generated, "a secret parameter is stored encrypted")
	assert.NotContains(t, data.Base.Rendered, generated, "the base holds no secret")
	password, err := data.Params["password"].Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, generated, password)
}

func TestCreateFromTemplateAsksForTheFirstDeployment(t *testing.T) {
	uc, fakes := newCreateTest(t)

	one := provisionOne(t, uc, fakes)

	assert.Equal(t, []string{"main-db"}, fakes.provision.deployFor)
	assert.Equal(t, base.DeploymentTriggerSourceAPI, one.Deployment.Trigger.Source)
	assert.Equal(t, "user-1", one.Deployment.Trigger.SourceID)
}

func TestCreateFromTemplateAuditsWithoutSecrets(t *testing.T) {
	uc, fakes := newCreateTest(t)

	provisionOne(t, uc, fakes)

	assert.Len(t, fakes.audit.entries, 1)
	entry := fakes.audit.entries[0]
	assert.Equal(t, base.AuditLogTypeAppCreate, entry.Type)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Contains(t, entry.Detail, `"template":"pg"`)
	assert.Contains(t, entry.Detail, `"revision":"0123456789abcdef0123456789abcdef01234567"`)
	assert.NotContains(t, entry.Detail, fakes.templates.resp.Result.Params["password"].Text())
}

func TestCreateFromTemplateRecordsAnImageOverride(t *testing.T) {
	uc, fakes := newCreateTest(t)
	rendered := fakes.templates.resp
	rendered.Result.ImageOverride = "postgres:18.7-alpine3.24"
	rendered.Result.ImageOverrideClass = templatemodel.ImageOverrideSameLine

	data := binding(t, provisionOne(t, uc, fakes).App)

	assert.Equal(t, "postgres:18.7-alpine3.24", data.ImageOverride)
	assert.Equal(t, "18", data.Version, "the version still describes the template, not the override")
	assert.Contains(t, fakes.audit.entries[0].Detail, `"imageOverride":"postgres:18.7-alpine3.24"`)
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

	provisioned, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, renderedWithDependency(t, fakes)))

	assert.NoError(t, err)
	assert.Equal(t, []string{"blog-db", "blog"}, fakes.provision.names, "the database first")
	db, blog := provisioned.Apps[0].App, provisioned.Apps[1].App
	assert.Equal(t, blog.ID, binding(t, db).CreatedForAppID)
	assert.Equal(t, db.ID, binding(t, blog).Dependencies[0].AppID)
	assert.Len(t, fakes.audit.entries, 2)
	assert.Contains(t, fakes.audit.entries[1].Detail, db.ID, "the app's creation names its database")
}

func TestProvisionAllNamesTheDependencyThatFailed(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.provision.failName = "blog-db"
	req := testCreateReq()
	req.Name = "blog"

	provisioned, err := uc.provisionAll(context.Background(), nil, testAuth(), req,
		planApps(req, renderedWithDependency(t, fakes)))

	assert.ErrorIs(t, err, errTestProvision)
	assert.Contains(t, errDetail(t, err), "blog-db")
	assert.Empty(t, provisioned.Apps)
	assert.Empty(t, fakes.audit.entries, "nothing is recorded for a creation that did not happen")
}

func TestTransformCreatedSeparatesTheAppFromItsDependencies(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "blog"
	apps := planApps(req, renderedWithDependency(t, fakes))

	provisioned, err := uc.provisionAll(context.Background(), nil, testAuth(), req, apps)
	assert.NoError(t, err)
	resp := transformCreated(apps, provisioned.Apps)

	assert.Equal(t, provisioned.Apps[1].App.ID, resp.Data.App.ID)
	assert.Equal(t, "dep-blog", resp.Data.Deployment.ID)
	assert.Len(t, resp.Data.Dependencies, 1)
	assert.Equal(t, "db", resp.Data.Dependencies[0].Name)
	assert.Equal(t, provisioned.Apps[0].App.ID, resp.Data.Dependencies[0].App.ID)
}
