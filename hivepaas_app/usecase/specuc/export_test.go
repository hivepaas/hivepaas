package specuc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

type fakePermissionManager struct {
	permission.Manager
	revealCalls int
	lastSubject *permission.RevealSubject
	denyReveal  bool
}

func (f *fakePermissionManager) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	f.revealCalls++
	f.lastSubject = subject
	if f.denyReveal {
		return hperrors.Wrap(hperrors.ErrForbidden)
	}
	return nil
}

type fakeSpecService struct {
	lastReq *specservice.ExportReq
}

func (f *fakeSpecService) Export(
	_ context.Context, _ database.IDB, req *specservice.ExportReq,
) (*specservice.ExportResp, error) {
	f.lastReq = req

	path := filepath.Join(req.WorkDir, "bundle.tar.gz")
	if err := os.WriteFile(path, []byte("archive"), 0o600); err != nil {
		return nil, err
	}
	return &specservice.ExportResp{
		Path: path, Filename: "hivepaas-spec.tar.gz", Size: 7, Report: &specmodel.Report{},
	}, nil
}

func newTestUC(t *testing.T) (*UC, *fakePermissionManager, *fakeSpecService) {
	t.Helper()
	perm := &fakePermissionManager{}
	svc := &fakeSpecService{}
	return New(nil, perm, svc), perm, svc
}

func testAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "u1", Role: base.UserRoleMember}}}
}

func exportReq(mode specmodel.SecretsMode) *specdto.ExportSpecReq {
	req := specdto.NewExportSpecReq()
	req.SecretsMode = mode
	return req
}

func TestExportSpecNeedsNoRevealForOmitMode(t *testing.T) {
	uc, perm, _ := newTestUC(t)
	perm.denyReveal = true

	resp, err := uc.ExportSpec(context.Background(), testAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())
	assert.Equal(t, 0, perm.revealCalls, "omit decrypts nothing")
}

func TestExportSpecRequiresRevealForBothSecretModes(t *testing.T) {
	for _, mode := range []specmodel.SecretsMode{
		specmodel.SecretsModePlaintext,
		specmodel.SecretsModeEncrypted,
	} {
		uc, perm, _ := newTestUC(t)
		perm.denyReveal = true

		req := exportReq(mode)
		req.Passphrase = "pw"

		_, err := uc.ExportSpec(context.Background(), testAuth(), req)
		assert.Error(t, err, "mode %v must be gated", mode)
		assert.Equal(t, 1, perm.revealCalls)
	}
}

// The subject is what the audit entry is written from, so it has to name the
// scope and say which mode was asked for.
func TestExportSpecRecordsAMeaningfulRevealSubject(t *testing.T) {
	uc, perm, _ := newTestUC(t)

	req := exportReq(specmodel.SecretsModePlaintext)
	req.ProjectID = "01JAB9XED0GTXBSQDFVYAJ8WB1"

	resp, err := uc.ExportSpec(context.Background(), testAuth(), req)
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())

	assert.Equal(t, base.ObjectScopeProject, perm.lastSubject.Scope)
	assert.Equal(t, "01JAB9XED0GTXBSQDFVYAJ8WB1", perm.lastSubject.ObjectID)
	assert.Contains(t, perm.lastSubject.ResName, "plaintext")
}

func TestExportSpecBuildsTheScopeFromTheRequest(t *testing.T) {
	uc, _, svc := newTestUC(t)

	req := exportReq(specmodel.SecretsModeOmit)
	req.ProjectID = "01JAB9XED0GTXBSQDFVYAJ8WB1"
	req.ProjectEnvID = "01JAB9XED0GTXBSQDFVYAJ8WB1:dev"
	req.AppID = "01JAB9XED0GTXBSQDFVYAJ8WD1"

	resp, err := uc.ExportSpec(context.Background(), testAuth(), req)
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())

	assert.Equal(t, base.ObjectScopeApp, svc.lastReq.Scope.ScopeType)
	assert.Equal(t, "01JAB9XED0GTXBSQDFVYAJ8WD1", svc.lastReq.Scope.AppID)
}

// The exporter stages a directory and archives it, so closing the body is the
// only moment the response is known to be finished with it.
func TestExportSpecRemovesTheWorkDirOnClose(t *testing.T) {
	uc, _, svc := newTestUC(t)

	resp, err := uc.ExportSpec(context.Background(), testAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.NoError(t, err)

	workDir := svc.lastReq.WorkDir
	assert.DirExists(t, workDir)

	assert.NoError(t, resp.Data.Content.Close())
	assert.NoDirExists(t, workDir, "the staging directory must not outlive the download")
}

func TestExportSpecSetsTheDownloadFilename(t *testing.T) {
	uc, _, _ := newTestUC(t)

	resp, err := uc.ExportSpec(context.Background(), testAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.NoError(t, err)
	defer resp.Data.Content.Close()

	assert.Equal(t, "application/gzip", resp.Data.ContentType)
	assert.Contains(t, resp.Data.ExtraHeaders["Content-Disposition"], "hivepaas-spec.tar.gz")
}
