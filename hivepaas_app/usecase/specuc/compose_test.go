package specuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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
}

func (f *fakeBundleService) PlanBundle(
	_ context.Context, _ database.IDB, req *specservice.PlanBundleReq,
) (*specmodel.ImportPlan, error) {
	f.planned = req
	return &specmodel.ImportPlan{PlanHash: "h4sh"}, nil
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
		FileName: "shop"}, resp.Data.Project)
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
