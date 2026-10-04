package specuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// fakeBundleService plans and applies a bundle built in memory, keeping what
// it was asked.
type fakeBundleService struct {
	fakeSpecService
	planned *specservice.PlanBundleReq
	applied *specservice.ApplyBundleReq
	// current is the existing env CurrentEnv answers.
	current *specmodel.EnvDoc
}

func (f *fakeBundleService) PlanBundle(
	_ context.Context, _ database.IDB, req *specservice.PlanBundleReq,
) (*specmodel.ImportPlan, error) {
	f.planned = req
	return &specmodel.ImportPlan{PlanHash: "h4sh"}, nil
}

func (f *fakeBundleService) CurrentEnv(
	_ context.Context, _ database.IDB, _, envKey string,
) (*specmodel.EnvDoc, error) {
	if f.current == nil || f.current.Env != envKey {
		return nil, hperrors.Wrap(hperrors.ErrProjectEnvNotFound)
	}
	return f.current, nil
}

func (f *fakeBundleService) ApplyBundle(
	_ context.Context, _ database.IDB, req *specservice.ApplyBundleReq,
) (*specservice.ApplyImportResp, error) {
	f.applied = req
	return applyResp(), nil
}

const shopCompose = `
name: shop
services:
  web:
    image: shop:1
    environment: {DB_PASSWORD: "${DB_PASSWORD:?}"}
`

func composeReq(compose string) *specdto.ValidateComposeReq {
	req := specdto.NewValidateComposeReq()
	req.Compose = compose
	_ = req.ModifyRequest()
	return req
}

func withPassword(req *specdto.ValidateComposeReq) *specdto.ValidateComposeReq {
	password := "pw"
	req.Variables = map[string]*specdto.ComposeVariableReq{"DB_PASSWORD": {Value: &password}}
	return req
}

// With no name of its own, the project takes the file's; its env is
// production.
func TestValidateComposeNamesTheProjectAfterTheFile(t *testing.T) {
	uc, _, _ := newTestUC(t)
	svc := &fakeBundleService{}
	uc.specService = svc

	resp, err := uc.ValidateCompose(context.Background(), adminAuth(), withPassword(composeReq(shopCompose)))

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, &specdto.ComposeProjectResp{Name: "shop", Key: "shop", Env: "production", EnvKey: "prod",
		NewEnv: true, FileName: "shop"}, resp.Data.Project)
	assert.Equal(t, "h4sh", resp.Data.Plan.PlanHash)
	if assert.NotNil(t, svc.planned) {
		assert.Contains(t, svc.planned.Doc.Envs["shop"], "prod")
		assert.Nil(t, svc.planned.MayMountSecrets, "the file's values are the request's own")
		assert.Nil(t, svc.planned.AuthorizeSecrets)
	}
}

func TestValidateComposeRefusesATakenName(t *testing.T) {
	uc, _, _ := newTestUC(t)
	uc.projectService = &fakeProjectService{taken: []string{"shop"}}

	_, err := uc.ValidateCompose(context.Background(), adminAuth(), withPassword(composeReq(shopCompose)))

	assert.ErrorIs(t, err, hperrors.ErrAlreadyExist)
}

// Until a required variable has a value there is nothing to plan: the review
// shows what is missing.
func TestValidateComposeWaitsForARequiredVariable(t *testing.T) {
	uc, _, _ := newTestUC(t)
	svc := &fakeBundleService{}
	uc.specService = svc

	resp, err := uc.ValidateCompose(context.Background(), adminAuth(), composeReq(shopCompose))

	assert.NoError(t, err)
	assert.Nil(t, resp.Data.Plan)
	assert.Nil(t, svc.planned)
	if assert.Len(t, resp.Data.Variables, 1) {
		assert.True(t, resp.Data.Variables[0].Required)
		assert.False(t, resp.Data.Variables[0].Given)
	}
}

func applyComposeReq(compose string) *specdto.ApplyComposeReq {
	req := specdto.NewApplyComposeReq()
	req.ValidateComposeReq = *withPassword(composeReq(compose))
	req.PlanHash, req.AcceptIssues, req.Deploy = "h4sh", true, true
	return req
}

// Applied, the project is recorded as created from a compose file, with what
// it was read into.
func TestApplyComposeIsRecorded(t *testing.T) {
	uc, audit, _ := newTestUC(t)
	svc := &fakeBundleService{}
	uc.specService = svc

	read, _, err := uc.applyComposeInTx(context.Background(), nil, adminAuth(), applyComposeReq(shopCompose))

	assert.NoError(t, err)
	assert.Equal(t, "shop", read.project.Key)
	if assert.NotNil(t, svc.applied) {
		assert.Equal(t, "h4sh", svc.applied.PlanHash)
		assert.True(t, svc.applied.Options.DeployCreated)
		assert.Equal(t, "usr_admin", svc.applied.OperatorID)
	}
	entries := entriesOfType(audit, base.AuditLogTypeComposeImport)
	if assert.Len(t, entries, 1) {
		for _, want := range []string{"shop", "production", "d1gest", `"create":2`} {
			assert.Contains(t, entries[0].Detail, want)
		}
	}
}

func TestApplyComposeRefusesWhileAVariableIsMissing(t *testing.T) {
	uc, _, _ := newTestUC(t)
	svc := &fakeBundleService{}
	uc.specService = svc
	req := applyComposeReq(shopCompose)
	req.Variables = nil

	_, _, err := uc.applyComposeInTx(context.Background(), nil, adminAuth(), req)

	assert.ErrorIs(t, err, hperrors.ErrSpecImportBlocked)
	assert.Nil(t, svc.applied)
}

// fakeShopProjectRepo has the project shop, with a production and a staging
// env.
type fakeShopProjectRepo struct {
	repository.ProjectRepo
}

func (f *fakeShopProjectRepo) GetByID(
	_ context.Context, _ database.IDB, id string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	if id != "prj_shop" {
		return nil, hperrors.Wrap(hperrors.ErrProjectNotFound)
	}
	return &entity.Project{ID: id, Key: "shop", Name: "Shop", Status: base.ProjectStatusActive,
		ProjectEnvs: []*entity.ProjectEnv{
			{Name: "production", Key: "prod", Color: "#f00", Index: 0},
			{Name: "staging", Key: "staging", Index: 1},
		}}, nil
}

func intoShop(env string, newEnv bool) *specdto.ValidateComposeReq {
	req := withPassword(composeReq(shopCompose))
	req.ProjectID = "prj_shop"
	req.Project = specdto.ComposeProjectReq{Name: "ignored", Env: env, NewEnv: newEnv, EnvColor: "#0f0"}
	return req
}

// Into an env the project has: planned at the env, keeping what is there, and
// the converter told of its apps - one the file names waits for the review.
func TestValidateComposeIntoAnExistingEnv(t *testing.T) {
	uc, _, _ := newTestUC(t)
	uc.projectRepo = &fakeShopProjectRepo{}
	svc := &fakeBundleService{current: &specmodel.EnvDoc{Env: "prod", Apps: map[string]*specmodel.AppDoc{
		"web": {App: "web"},
	}}}
	uc.specService = svc

	resp, err := uc.ValidateCompose(context.Background(), adminAuth(), intoShop("production", false))

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, &specdto.ComposeProjectResp{ID: "prj_shop", Name: "Shop", Key: "shop", Env: "production",
		EnvKey: "prod", FileName: "shop"}, resp.Data.Project)
	if assert.NotNil(t, svc.planned) {
		assert.Equal(t, base.ObjectScopeProjectEnv, svc.planned.Scope.ScopeType)
		assert.Equal(t, specmodel.ExistingKeep, svc.planned.Options.Existing)
		assert.Nil(t, svc.planned.Doc.Projects["shop"].Owner, "the project's owner is not changed")
		assert.Equal(t, "#f00", svc.planned.Doc.Envs["shop"]["prod"].Color)
		issues := svc.planned.Issues["projects/shop/envs/prod/apps/web"]
		if assert.NotEmpty(t, issues) {
			assert.Equal(t, composeservice.CodeAppExists, issues[0].Code)
		}
	}
	assert.Equal(t, "web", resp.Data.Services[0].Existing)
}

// A new env of the project comes after its others, planned at the project.
func TestValidateComposeIntoANewEnv(t *testing.T) {
	uc, _, _ := newTestUC(t)
	uc.projectRepo = &fakeShopProjectRepo{}
	svc := &fakeBundleService{}
	uc.specService = svc

	resp, err := uc.ValidateCompose(context.Background(), adminAuth(), intoShop("qa", true))

	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, resp.Data.Project.NewEnv)
	if assert.NotNil(t, svc.planned) {
		assert.Equal(t, base.ObjectScopeProject, svc.planned.Scope.ScopeType)
		env := svc.planned.Doc.Envs["shop"]["qa"]
		if assert.NotNil(t, env) {
			assert.Equal(t, 2, env.Index)
			assert.Equal(t, "#0f0", env.Color)
		}
	}
}

func TestValidateComposeRefusesANewEnvTheProjectHas(t *testing.T) {
	uc, _, _ := newTestUC(t)
	uc.projectRepo = &fakeShopProjectRepo{}
	uc.specService = &fakeBundleService{}

	_, err := uc.ValidateCompose(context.Background(), adminAuth(), intoShop("prod", true))
	assert.ErrorContains(t, err, "has an env named 'prod' already", "by its key")

	_, err = uc.ValidateCompose(context.Background(), adminAuth(), intoShop("qa", false))
	assert.ErrorIs(t, err, hperrors.ErrProjectEnvNotFound)
}

// Applied into a project, the entry is the project's.
func TestApplyComposeIntoAProjectIsRecordedOnIt(t *testing.T) {
	uc, audit, _ := newTestUC(t)
	uc.projectRepo = &fakeShopProjectRepo{}
	uc.specService = &fakeBundleService{}
	req := applyComposeReq(shopCompose)
	req.ValidateComposeReq = *intoShop("qa", true)

	_, _, err := uc.applyComposeInTx(context.Background(), nil, adminAuth(), req)

	assert.NoError(t, err)
	entries := entriesOfType(audit, base.AuditLogTypeComposeImport)
	if assert.Len(t, entries, 1) {
		assert.Equal(t, base.ObjectScopeProject, entries[0].Scope)
		assert.Equal(t, "prj_shop", entries[0].ObjectID)
		assert.Contains(t, entries[0].Detail, `"newEnv":true`)
	}
}
