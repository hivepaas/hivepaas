package specuc

import (
	"context"
	"errors"
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
	// failType makes Record fail for one entry type, to test what an export does
	// when its record cannot be written.
	failType base.AuditLogType
}

func (f *fakeAuditService) Record(
	_ context.Context, _ database.IDB, entry *auditservice.Entry,
) error {
	if f.failType != "" && entry.Type == f.failType {
		return errors.New("audit store unavailable")
	}
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
	specservice.Service
	lastReq *specservice.ExportReq
	err     error
}

func (f *fakeSpecService) Export(
	_ context.Context, _ database.IDB, req *specservice.ExportReq,
) (*specservice.ExportResp, error) {
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}

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
	return New(nil, manager, audit, svc), audit, svc
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

// exportReq builds a global-scope request. The handler resolves the scope from
// the route before the usecase runs, so a request always arrives with one.
func exportReq(mode specmodel.SecretsMode) *specdto.ExportSpecReq {
	req := specdto.NewExportSpecReq()
	req.Scope = entity.NewObjectScopeGlobal()
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

	reveals := entriesOfType(audit, base.AuditLogTypeSecretReveal)
	assert.Len(t, reveals, 1)
	assert.Equal(t, base.AuditLogResultAllowed, reveals[0].Result)
	assert.Len(t, entriesOfType(audit, base.AuditLogTypeSpecExport), 1)
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
	assert.Empty(t, entriesOfType(audit, base.AuditLogTypeSecretReveal),
		"nothing was revealed, so no reveal is recorded")
	assert.Len(t, entriesOfType(audit, base.AuditLogTypeSpecExport), 1,
		"but the export itself still is")
}

// The usecase no longer derives the scope - the handler does - so what matters
// here is that it reaches the exporter unchanged.
func TestExportSpecPassesTheScopeThroughToTheExporter(t *testing.T) {
	uc, _, svc := newTestUC(t)

	req := exportReq(specmodel.SecretsModeOmit)
	req.Scope = entity.NewObjectScopeApp("01JAB9XED0GTXBSQDFVYAJ8WD1", "",
		"01JAB9XED0GTXBSQDFVYAJ8WB1", "01JAB9XED0GTXBSQDFVYAJ8WB1:dev")

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
		"encrypted without passphrase": {
			Scope:       entity.NewObjectScopeGlobal(),
			SecretsMode: specmodel.SecretsModeEncrypted,
		},
		"unknown mode": {
			Scope:       entity.NewObjectScopeGlobal(),
			SecretsMode: specmodel.SecretsMode("none"),
		},
	} {
		uc, audit, svc := newTestUC(t)

		_, err := uc.ExportSpec(context.Background(), adminAuth(), req)
		assert.Error(t, err, name)
		assert.Empty(t, audit.entries, "%s: nothing was revealed, so nothing may be recorded", name)
		assert.Nil(t, svc.lastReq, name)
	}
}

// The audit entry is written from the reveal subject, so it has to name the
// scope the request carried - now that the handler, not the usecase, builds it.
func TestExportSpecNamesTheRequestScopeInTheRevealAudit(t *testing.T) {
	allowSecretReveal(t, true)
	uc, audit, _ := newTestUC(t)

	req := exportReq(specmodel.SecretsModePlaintext)
	req.Scope = entity.NewObjectScopeProject("01JAB9XED0GTXBSQDFVYAJ8WB1")

	resp, err := uc.ExportSpec(context.Background(), adminAuth(), req)
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())

	reveals := entriesOfType(audit, base.AuditLogTypeSecretReveal)
	assert.Len(t, reveals, 1)
	assert.Equal(t, base.ObjectScopeProject, reveals[0].Scope)
	assert.Equal(t, "01JAB9XED0GTXBSQDFVYAJ8WB1", reveals[0].ObjectID)
	assert.Contains(t, reveals[0].ResName, "plaintext")
}

// entriesOfType picks one kind of entry out of everything recorded, since an
// export that reveals secrets writes two.
func entriesOfType(audit *fakeAuditService, typ base.AuditLogType) []*auditservice.Entry {
	var out []*auditservice.Entry
	for _, entry := range audit.entries {
		if entry.Type == typ {
			out = append(out, entry)
		}
	}
	return out
}

// Every export is recorded, whatever the secrets mode. One with every secret
// emptied still hands over the full shape of what it covers.
func TestExportSpecRecordsTheExportInEveryMode(t *testing.T) {
	allowSecretReveal(t, true)

	for _, mode := range []specmodel.SecretsMode{
		specmodel.SecretsModeOmit,
		specmodel.SecretsModeEncrypted,
		specmodel.SecretsModePlaintext,
	} {
		uc, audit, _ := newTestUC(t)

		req := exportReq(mode)
		req.Scope = entity.NewObjectScopeApp("01JAB9XED0GTXBSQDFVYAJ8WD1", "",
			"01JAB9XED0GTXBSQDFVYAJ8WB1", "dev")

		resp, err := uc.ExportSpec(context.Background(), adminAuth(), req)
		assert.NoError(t, err, mode)
		assert.NoError(t, resp.Data.Content.Close())

		exports := entriesOfType(audit, base.AuditLogTypeSpecExport)
		assert.Len(t, exports, 1, mode)
		entry := exports[0]

		assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
		assert.Equal(t, base.AuditLogSourceAPIAction, entry.Source)
		assert.Equal(t, base.ObjectScopeApp, entry.Scope)
		assert.Equal(t, "01JAB9XED0GTXBSQDFVYAJ8WD1", entry.ObjectID)
		assert.Equal(t, base.ResourceTypeApp, entry.ResType)
		assert.Contains(t, entry.Detail, `"secretsMode":"`+string(mode)+`"`)
		assert.Contains(t, entry.Detail, "hivepaas-spec.tar.gz")
	}
}

// The passphrase must not reach the audit log any more than an access log.
func TestExportSpecNeverRecordsThePassphrase(t *testing.T) {
	allowSecretReveal(t, true)
	uc, audit, _ := newTestUC(t)

	req := exportReq(specmodel.SecretsModeEncrypted)
	req.Passphrase = "sentinel-passphrase-Zq9"

	resp, err := uc.ExportSpec(context.Background(), adminAuth(), req)
	assert.NoError(t, err)
	assert.NoError(t, resp.Data.Content.Close())

	assert.NotEmpty(t, audit.entries)
	for _, entry := range audit.entries {
		assert.NotContains(t, entry.Detail, "sentinel-passphrase-Zq9")
		assert.NotContains(t, entry.ResName, "sentinel-passphrase-Zq9")
	}
}

// An export whose record cannot be written is not handed over, and its staging
// directory does not linger: delivering the archive without the record would
// leave exactly the gap the record exists to close.
func TestExportSpecIsAbortedWhenItCannotBeRecorded(t *testing.T) {
	uc, audit, svc := newTestUC(t)
	audit.failType = base.AuditLogTypeSpecExport

	resp, err := uc.ExportSpec(context.Background(), plainAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.Error(t, err)
	assert.Nil(t, resp)

	assert.NotNil(t, svc.lastReq, "precondition: the bundle was built")
	assert.NoDirExists(t, svc.lastReq.WorkDir, "and its staging directory was removed")
}

// A failed export took nothing away, so there is nothing to record.
func TestExportSpecRecordsNothingWhenTheExportFails(t *testing.T) {
	uc, audit, svc := newTestUC(t)
	svc.err = errors.New("walk failed")

	_, err := uc.ExportSpec(context.Background(), plainAuth(), exportReq(specmodel.SecretsModeOmit))
	assert.Error(t, err)
	assert.Empty(t, entriesOfType(audit, base.AuditLogTypeSpecExport))
}
