package specserviceimpl

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice/composeserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const composeStack = `
services:
  db:
    image: postgres:17
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes: [pgdata:/var/lib/postgresql/data]
  app:
    image: ghcr.io/me/app:1.2
    command: ["sh", "-c", "migrate && serve"]
    depends_on: [db]
    ports: ["8080:80"]
    environment:
      DATABASE_URL: postgres://app:${DB_PASSWORD}@db/app
      API_TOKEN: t0ken
    secrets: [api_key]
    configs: [{source: settings, target: /etc/app/settings.json}]
    volumes: [pgdata:/backup:ro]
volumes:
  pgdata:
secrets:
  api_key: {environment: API_KEY}
configs:
  settings: {content: '{"a": 1}'}
`

// What a compose file reads into, the import plans and writes with nothing
// missing: the project's default volume, the env's secrets and config files
// the apps mount and refer to, the owner of a shared volume first.
func TestAComposeFilesBundlePlansAndApplies(t *testing.T) {
	svc, _ := planFixture(t)
	svc.appProvisionService.(*fakeProvisionService).repo = svc.appRepo.(*fakeAppRepo)
	converted, err := composeserviceimpl.New().Convert(context.Background(), &composeservice.ConvertReq{
		Compose: composeStack, DotEnv: "DB_PASSWORD=pw\nAPI_KEY=k\n",
		ProjectKey: "blog", ProjectName: "Blog", EnvKey: "prod", EnvName: "production",
		NetworkName: "blog_prod_net", RootDomain: "example.com",
	})
	if !assert.NoError(t, err) {
		return
	}
	req := &specservice.PlanBundleReq{
		ValidateImportReq: specservice.ValidateImportReq{Scope: entity.NewObjectScopeGlobal(),
			Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate, DeployCreated: true}},
		Doc: converted.Bundle, Issues: converted.Issues,
	}
	plan, err := svc.PlanBundle(context.Background(), nil, req)
	if !assert.NoError(t, err) {
		return
	}
	for _, n := range plan.Nodes {
		for _, issue := range n.Issues {
			assert.True(t, strings.HasPrefix(issue.Code, "COMPOSE_"), "%s: %s %v", n.Path, issue.Code, issue.Detail)
		}
	}
	assert.Equal(t, specmodel.ActionCreate, node(t, plan, "projects/blog/envs/prod/apps/app").Action)
	assert.True(t, node(t, plan, "projects/blog/envs/prod/apps/app").Deploy)

	resp, err := svc.ApplyBundle(context.Background(), nil, &specservice.ApplyBundleReq{
		PlanBundleReq: *req, OperatorID: "u_operator", PlanHash: plan.PlanHash, AcceptIssues: true,
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, resp.Deployments, 2)
	// Phase 2 builds the environment of the env whose secrets were written: it
	// panicked on a scope of ids alone.
	assert.NoError(t, resp.AfterCommit(context.Background(), nil))

	provision := svc.appProvisionService.(*fakeProvisionService)
	if assert.Len(t, provision.reqs, 2) {
		assert.Equal(t, "db", provision.reqs[0].Key, "the volume's owner first")
	}
	var app string
	for _, one := range provision.reqs {
		if one.Key == "app" {
			app = one.AppID
		}
	}
	types := map[base.SettingType]*entity.Setting{}
	mounted := 0
	for _, setting := range provision.settings[app] {
		types[setting.Type] = setting
		if setting.Type == base.SettingTypeAppSettingMount {
			mounted++
			assert.True(t, settingmountservice.ValidEntryKey(setting.Name),
				"an entry called %q is never mounted", setting.Name)
		}
	}
	assert.Equal(t, 2, mounted, "the secret's and the config's")
	for _, typ := range []base.SettingType{base.SettingTypeEnvVar, base.SettingTypeAppRouting,
		base.SettingTypeAppSettingMount, base.SettingTypeAppDeployment} {
		assert.Contains(t, types, typ)
	}
	assert.Contains(t, types[base.SettingTypeEnvVar].Data, "${secrets.DB_PASSWORD}")
	assert.Contains(t, types[base.SettingTypeEnvVar].Data, "${secrets.API_TOKEN}")
	if assert.Contains(t, types, base.SettingTypeSecret, "the app's own secret") {
		assert.Equal(t, "API_TOKEN", types[base.SettingTypeSecret].Name)
	}
	assert.Equal(t, "sh -c 'migrate && serve'",
		types[base.SettingTypeAppDeployment].MustAsAppDeploymentSettings().Command)

	secrets := 0
	for _, setting := range persisted(svc).UpsertingSettings {
		if setting.Type == base.SettingTypeSecret && setting.Scope == base.ObjectScopeProjectEnv {
			secrets++
			assert.True(t, setting.Inheritable, setting.Name)
		}
	}
	assert.Equal(t, 2, secrets, "the variable's and the file's")
}

const composeIntoDev = `
services:
  backend:
    image: ghcr.io/me/backend:2
  worker:
    image: ghcr.io/me/worker:2
    depends_on: [backend]
`

// Into an env that has an app the file names: blocked until the review
// chooses, then the app there is kept as it is and the rest created.
func TestAComposeFileIntoAnExistingEnvKeepsWhatItHas(t *testing.T) {
	svc, _ := planFixture(t)
	current, err := svc.CurrentEnv(context.Background(), nil, "p1", "dev")
	if !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, current.Apps, "backend")

	planWith := func(services map[string]*composeservice.ServiceReq) *specmodel.ImportPlan {
		t.Helper()
		converted, convertErr := composeserviceimpl.New().Convert(context.Background(), &composeservice.ConvertReq{
			Compose: composeIntoDev, ProjectKey: "project_a", ProjectName: "Project A", EnvKey: "dev",
			EnvName: "dev", NetworkName: "project_a_dev_net", Existing: current, Services: services,
		})
		if !assert.NoError(t, convertErr) {
			t.FailNow()
		}
		plan, planErr := svc.PlanBundle(context.Background(), nil, &specservice.PlanBundleReq{
			ValidateImportReq: specservice.ValidateImportReq{Scope: entity.NewObjectScopeProjectEnv("p1", "dev"),
				Options: specmodel.ImportOptions{Existing: specmodel.ExistingKeep}},
			Doc: converted.Bundle, Issues: converted.Issues,
		})
		if !assert.NoError(t, planErr) {
			t.FailNow()
		}
		return plan
	}

	plan := planWith(nil)
	assert.Positive(t, plan.Summary[string(specmodel.SeverityBlocked)])
	backend := node(t, plan, "projects/project_a/envs/dev/apps/backend")
	if assert.NotEmpty(t, backend.Issues) {
		assert.Equal(t, composeservice.CodeAppExists, backend.Issues[0].Code)
	}

	plan = planWith(map[string]*composeservice.ServiceReq{"backend": {UseExisting: true}})
	assert.Zero(t, plan.Summary[string(specmodel.SeverityBlocked)])
	assert.Equal(t, specmodel.ActionKeep, node(t, plan, "projects/project_a/envs/dev/apps/backend").Action)
	assert.Equal(t, specmodel.ActionCreate, node(t, plan, "projects/project_a/envs/dev/apps/worker").Action)
	for _, n := range plan.Nodes {
		if n.Action == specmodel.ActionUpdate {
			t.Errorf("%s is changed: %v", n.Path, n.Changes)
		}
	}
}

// A YAML "\0" is a NUL once read: a Compose file holding one is refused as a
// bundle would be, saying where, rather than failing as it is written.
func TestAComposeFileHoldingANULIsRefused(t *testing.T) {
	svc, _ := planFixture(t)
	converted, err := composeserviceimpl.New().Convert(context.Background(), &composeservice.ConvertReq{
		Compose:    "services:\n  app:\n    image: nginx\n    environment:\n      GREETING: \"hi\\0there\"\n",
		ProjectKey: "blog", ProjectName: "Blog", EnvKey: "prod", EnvName: "production",
		NetworkName: "blog_prod_net", RootDomain: "example.com",
	})
	if !assert.NoError(t, err) {
		return
	}

	_, err = svc.PlanBundle(context.Background(), nil, &specservice.PlanBundleReq{
		ValidateImportReq: specservice.ValidateImportReq{Scope: entity.NewObjectScopeGlobal(),
			Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate}},
		Doc: converted.Bundle, Issues: converted.Issues,
	})

	assert.ErrorIs(t, err, hperrors.ErrSpecBundleInvalid)
}
