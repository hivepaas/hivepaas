package specuc

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission/permissionimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// The permission manager is the real one. A double would only prove the export
// calls something - not that a user without the capability is actually refused,
// which is the property worth having.
//
// Only the audit service and the ACL repository are faked, the way
// permissionimpl's own tests do it: embedding the interface means a method this
// path does not reach panics rather than silently passing.

type fakeAuditService struct {
	auditservice.Service
	entries []*auditservice.Entry
}

func (f *fakeAuditService) Record(
	_ context.Context, _ database.IDB, entry *auditservice.Entry,
) error {
	f.entries = append(f.entries, entry)
	return nil
}

// fakeACLRepo grants nothing, so the capability has to come from the role.
type fakeACLRepo struct {
	repository.ACLPermissionRepo
}

func (f *fakeACLRepo) ListByResources(
	_ context.Context, _ database.IDB, _ []*base.PermissionResource, _ ...bunex.SelectQueryOption,
) ([]*entity.ACLPermission, error) {
	return nil, nil
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

func newTestUC(t *testing.T) (*UC, *fakeAuditService, *fakeSpecService) {
	t.Helper()
	audit := &fakeAuditService{}
	manager := permissionimpl.NewManager(&fakeACLRepo{}, nil, nil, nil, audit)
	svc := &fakeSpecService{}
	return New(nil, manager, svc), audit, svc
}

// allowSecretReveal sets the operator flag that gates every stored secret.
// Without it even an admin is refused.
func allowSecretReveal(t *testing.T, enabled bool) {
	t.Helper()
	prev := config.Current()
	config.SetCurrent(&config.Config{Security: config.Security{ReturnSecretsViaAPI: enabled}})
	t.Cleanup(func() { config.SetCurrent(prev) })
}

func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{
		ID: "usr_admin", Role: base.UserRoleAdmin,
	}}}
}

func plainAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{
		ID: "usr_1", Role: base.UserRoleMember,
	}}}
}

func exportReq(mode specmodel.SecretsMode) *specdto.ExportSpecReq {
	req := specdto.NewExportSpecReq()
	req.SecretsMode = mode
	req.Passphrase = "correct horse battery staple"
	return req
}

// The question this answers: does a user without the reveal capability actually
// get refused, or is the gate only being called?
func TestExportSpecRefusesAUserWithoutTheRevealCapability(t *testing.T) {
	allowSecretReveal(t, true)

	for _, mode := range []specmodel.SecretsMode{
		specmodel.SecretsModePlaintext,
		specmodel.SecretsModeEncrypted,
	} {
		uc, audit, svc := newTestUC(t)

		_, err := uc.ExportSpec(context.Background(), plainAuth(), exportReq(mode))
		assert.Error(t, err, "mode %v must refuse a user without cap::secret::reveal", mode)
		assert.Nil(t, svc.lastReq, "and must not reach the exporter at all")

		assert.Len(t, audit.entries, 1, "the refusal is recorded")
		assert.Equal(t, base.AuditLogResultDenied, audit.entries[0].Result)
	}
}

func TestExportSpecAllowsAUserWhoHasTheCapability(t *testing.T) {
	allowSecretReveal(t, true)
	uc, audit, svc := newTestUC(t)

	resp, err := uc.ExportSpec(context.Background(), adminAuth(),
		exportReq(specmodel.SecretsModePlaintext))
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())
	assert.NotNil(t, svc.lastReq)

	assert.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogResultAllowed, audit.entries[0].Result)
}

// The operator flag outranks the capability: with secrets disabled system-wide,
// even an admin is refused.
func TestExportSpecHonorsTheOperatorSecretsFlag(t *testing.T) {
	allowSecretReveal(t, false)
	uc, _, svc := newTestUC(t)

	_, err := uc.ExportSpec(context.Background(), adminAuth(),
		exportReq(specmodel.SecretsModePlaintext))
	assert.Error(t, err)
	assert.Nil(t, svc.lastReq)
}

// omit reads nothing, so it is ungated - which is what makes it a safe default.
func TestExportSpecNeedsNoCapabilityForOmitMode(t *testing.T) {
	allowSecretReveal(t, false)
	uc, audit, svc := newTestUC(t)

	resp, err := uc.ExportSpec(context.Background(), plainAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())

	assert.NotNil(t, svc.lastReq)
	assert.Empty(t, audit.entries, "nothing was revealed, so there is nothing to record")
}

func TestExportSpecBuildsTheScopeFromTheRequest(t *testing.T) {
	uc, _, svc := newTestUC(t)

	req := exportReq(specmodel.SecretsModeOmit)
	req.ProjectID = "01JAB9XED0GTXBSQDFVYAJ8WB1"
	req.ProjectEnvID = "01JAB9XED0GTXBSQDFVYAJ8WB1:dev"
	req.AppID = "01JAB9XED0GTXBSQDFVYAJ8WD1"

	resp, err := uc.ExportSpec(context.Background(), plainAuth(), req)
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())

	assert.Equal(t, base.ObjectScopeApp, svc.lastReq.Scope.ScopeType)
	assert.Equal(t, "01JAB9XED0GTXBSQDFVYAJ8WD1", svc.lastReq.Scope.AppID)
}

// The exporter stages a directory and archives it, so closing the body is the
// only moment the response is known to be finished with it.
func TestExportSpecRemovesTheWorkDirOnClose(t *testing.T) {
	uc, _, svc := newTestUC(t)

	resp, err := uc.ExportSpec(context.Background(), plainAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.NoError(t, err)

	workDir := svc.lastReq.WorkDir
	assert.DirExists(t, workDir)

	assert.NoError(t, resp.Data.Content.Close())
	assert.NoDirExists(t, workDir, "the staging directory must not outlive the download")
}

func TestExportSpecSetsTheDownloadFilename(t *testing.T) {
	uc, _, _ := newTestUC(t)

	resp, err := uc.ExportSpec(context.Background(), plainAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.NoError(t, err)
	defer resp.Data.Content.Close()

	assert.Equal(t, "application/gzip", resp.Data.ContentType)
	assert.Contains(t, resp.Data.ExtraHeaders["Content-Disposition"], "hivepaas-spec.tar.gz")
}

// The passphrase must never be reachable as a query parameter. The access log
// records the full path including the query - a running instance logged
// `path=/_/spec/export?secretsMode=omit` - and this installation's own logging
// subsystem ships those lines to a searchable store.
func TestExportSpecReqCarriesThePassphraseInTheBodyOnly(t *testing.T) {
	field, ok := reflect.TypeFor[specdto.ExportSpecReq]().FieldByName("Passphrase")
	assert.True(t, ok)

	assert.Equal(t, "passphrase", field.Tag.Get("json"),
		"the passphrase is read from the JSON body")
	assert.Empty(t, field.Tag.Get("mapstructure"),
		"a mapstructure tag would make it bindable from the query string, which the access log records")
}

// Found against a running instance: an encrypted export with no passphrase was
// refused, yet the audit log already said "secret-reveal: allowed", because the
// gate ran before the request was checked. An audit trail must not claim
// secrets were released when nothing was.
func TestExportSpecRecordsNoRevealForARequestThatWillBeRefused(t *testing.T) {
	allowSecretReveal(t, true)

	for name, req := range map[string]*specdto.ExportSpecReq{
		"encrypted without passphrase": {SecretsMode: specmodel.SecretsModeEncrypted},
		"unknown mode":                 {SecretsMode: specmodel.SecretsMode("none")},
	} {
		uc, audit, svc := newTestUC(t)

		_, err := uc.ExportSpec(context.Background(), adminAuth(), req)
		assert.Error(t, err, name)
		assert.Empty(t, audit.entries, "%s: nothing was revealed, so nothing may be recorded", name)
		assert.Nil(t, svc.lastReq, name)
	}
}
