package specuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission/permissionimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// fakeImportService reads a bundle carrying the given mode, asks the gate the
// way the real one does, and plans nothing.
type fakeImportService struct {
	fakeSpecService
	mode    specmodel.SecretsMode
	planned bool
	lastReq *specservice.ValidateImportReq
}

func (f *fakeImportService) ValidateImport(
	ctx context.Context, _ database.IDB, req *specservice.ValidateImportReq,
) (*specservice.ValidateImportResp, error) {
	if f.mode.RevealsSecrets() && req.AuthorizeSecrets != nil {
		if err := req.AuthorizeSecrets(ctx, f.mode); err != nil {
			return nil, err
		}
	}
	f.planned = true
	f.lastReq = req
	return &specservice.ValidateImportResp{Plan: &specmodel.ImportPlan{}, SecretsMode: f.mode}, nil
}

func importReq() *specdto.ValidateImportReq {
	req := specdto.NewValidateImportReq()
	req.Scope = entity.NewObjectScopeGlobal()
	req.Bundle = []byte("bundle")
	req.ModifyRequest()
	return req
}

// Planning a bundle that carries secrets compares this installation's secrets
// with it, so it takes what exporting one takes.
func TestValidateImportRefusesSecretsToAUserWithoutTheRevealCapability(t *testing.T) {
	allowSecretReveal(t, true)
	uc, audit, _ := newTestUC(t)
	svc := &fakeImportService{mode: specmodel.SecretsModeEncrypted}
	uc.specService = svc

	_, err := uc.ValidateImport(context.Background(), plainAuth(), importReq())

	assert.Error(t, err)
	assert.False(t, svc.planned, "nothing is compared")
	assert.Len(t, audit.entries, 1, "the refusal is recorded")
}

func TestValidateImportAllowsSecretsToAUserWithTheCapability(t *testing.T) {
	allowSecretReveal(t, true)
	uc, _, _ := newTestUC(t)
	svc := &fakeImportService{mode: specmodel.SecretsModePlaintext}
	uc.specService = svc

	_, err := uc.ValidateImport(context.Background(), adminAuth(), importReq())

	assert.NoError(t, err)
	assert.True(t, svc.planned)
}

func TestValidateImportNeedsNoCapabilityWithoutSecrets(t *testing.T) {
	uc, audit, _ := newTestUC(t)
	svc := &fakeImportService{mode: specmodel.SecretsModeOmit}
	uc.specService = svc

	_, err := uc.ValidateImport(context.Background(), plainAuth(), importReq())

	assert.NoError(t, err)
	assert.Empty(t, audit.entries)
}

func TestValidateImportReqDefaultsToUpdatingWhatExists(t *testing.T) {
	assert.Equal(t, specmodel.ExistingUpdate, importReq().Options.Existing)
	req := importReq()
	req.Bundle = nil
	assert.NotEmpty(t, req.Validate(), "a request without a bundle is refused")
}

// fakeOwnerlessProjectRepo owns nothing for anybody, so an app's access has to
// come from the role.
type fakeOwnerlessProjectRepo struct {
	repository.ProjectRepo
}

func (f *fakeOwnerlessProjectRepo) GetByIDAndOwner(
	_ context.Context, _ database.IDB, _, _ string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	return nil, hperrors.Wrap(hperrors.ErrProjectNotFound)
}

// Granting capabilities and reaching another app's storage take what they take
// when a template asks for them.
func TestValidateImportAsksThePermissionManagerForWhatAnImportGrants(t *testing.T) {
	app := &entity.App{ID: "app_1", ProjectID: "p1", ProjectEnvID: "p1:dev"}
	for auth, want := range map[*basedto.Auth]bool{adminAuth(): true, plainAuth(): false} {
		uc, audit, _ := newTestUC(t)
		uc.permissionManager = permissionimpl.NewManager(&fakeACLRepo{}, nil, nil, &fakeOwnerlessProjectRepo{}, audit)
		svc := &fakeImportService{mode: specmodel.SecretsModeOmit}
		uc.specService = svc

		_, err := uc.ValidateImport(context.Background(), auth, importReq())
		assert.NoError(t, err)

		mayGrant, err := svc.lastReq.MayGrantCapabilities(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, want, mayGrant, auth.User.ID)
		mayWrite, err := svc.lastReq.MayWriteApp(context.Background(), app)
		assert.NoError(t, err)
		assert.Equal(t, want, mayWrite, auth.User.ID)
	}
}

// Changing an existing project's owner takes what project update takes.
func TestValidateImportAsksWhoMayChangeAProjectsOwner(t *testing.T) {
	for name, tc := range map[string]struct {
		auth  *basedto.Auth
		owner string
		want  bool
	}{
		"an admin":          {adminAuth(), "usr_x", true},
		"the current owner": {plainAuth(), "usr_1", true},
		"another member":    {plainAuth(), "usr_x", false},
	} {
		t.Run(name, func(t *testing.T) {
			uc, _, _ := newTestUC(t)
			svc := &fakeImportService{mode: specmodel.SecretsModeOmit}
			uc.specService = svc

			_, err := uc.ValidateImport(context.Background(), tc.auth, importReq())
			assert.NoError(t, err)

			allowed, err := svc.lastReq.MayChangeOwner(context.Background(), &entity.Project{ID: "p1", OwnerID: tc.owner})
			assert.NoError(t, err)
			assert.Equal(t, tc.want, allowed)
		})
	}
}
