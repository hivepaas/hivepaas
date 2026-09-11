package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

/// Fakes

type fakeAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (f *fakeAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return f.err
}

type fakePermissionManager struct {
	permission.Manager
	granted bool
	err     error
}

func (f *fakePermissionManager) CheckAccess(_ context.Context, _ database.IDB,
	_ *basedto.Auth, _ permission.AccessCheck) (bool, error) {
	return f.granted, f.err
}

/// Helpers

func newRevealUC(t *testing.T, granted bool, auditErr error) (*BaseUC, *fakeAuditService) {
	t.Helper()
	audit := &fakeAuditService{err: auditErr}
	return &BaseUC{
		AuditService:      audit,
		PermissionManager: &fakePermissionManager{granted: granted},
	}, audit
}

// newBasicAuthSetting builds a stored setting whose password is encrypted, the
// state a setting is in when it comes back from the database.
func newBasicAuthSetting(t *testing.T) *entity.Setting {
	t.Helper()
	key, err := datakey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	datakey.SetActive(key)

	setting := &entity.Setting{
		ID:   "set_1",
		Name: "web-auth",
		Type: base.SettingTypeBasicAuth,
	}
	if err := setting.SetData(&entity.BasicAuth{
		Username: "admin",
		Password: entity.NewEncryptedField("the-real-password"),
	}); err != nil {
		t.Fatal(err)
	}
	// Round-trip through the stored form so the field holds ciphertext only.
	stored := &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
	stored.ObjectID, stored.CurrentObjectID = "obj_1", "obj_1"
	return stored
}

func storedPassword(t *testing.T, setting *entity.Setting) *entity.EncryptedField {
	t.Helper()
	basicAuth, err := setting.AsBasicAuth()
	if err != nil {
		t.Fatal(err)
	}
	return &basicAuth.Password
}

func enableReveal(t *testing.T, enabled bool) {
	t.Helper()
	prev := config.Current()
	config.SetCurrent(&config.Config{Security: config.Security{ReturnSecretsViaAPI: enabled}})
	t.Cleanup(func() { config.SetCurrent(prev) })
}

// revealScope is the scope the reveal is being asked for. It is what the audit
// entry is filed under, and an entry filed under the wrong scope is invisible to
// the scoped audit log listing - which is where somebody would go looking for it.
var revealScope = &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "obj_1"}

/// Tests

func TestRevealSecretsAllowed(t *testing.T) {
	enableReveal(t, true)
	uc, audit := newRevealUC(t, true, nil)
	setting := newBasicAuthSetting(t)

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting); err != nil {
		t.Fatalf("reveal must succeed: %v", err)
	}
	if got := storedPassword(t, setting).String(); got != "the-real-password" {
		t.Errorf("secret was not revealed, got %q", got)
	}
	if len(audit.entries) != 1 || audit.entries[0].Result != base.AuditLogResultAllowed {
		t.Fatalf("expected one allowed entry, got %+v", audit.entries)
	}
	if audit.entries[0].ResID != "set_1" || audit.entries[0].ResName != "web-auth" {
		t.Error("the entry must name the setting it is about")
	}
	if audit.entries[0].Scope != base.ObjectScopeApp || audit.entries[0].ObjectID != "obj_1" {
		t.Errorf("the entry must be filed under the scope the reveal was asked for, got %v/%q",
			audit.entries[0].Scope, audit.entries[0].ObjectID)
	}
}

// A refusal is the more interesting half of the record: it is the only sign that
// somebody is trying doors.
func TestRevealSecretsDeniedIsStillRecorded(t *testing.T) {
	enableReveal(t, true)
	uc, audit := newRevealUC(t, false, nil)
	setting := newBasicAuthSetting(t)

	err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)
	if err == nil {
		t.Fatal("a caller without the permission must be refused")
	}
	if storedPassword(t, setting).IsEncrypted() != true {
		t.Error("the secret must stay encrypted when the caller is refused")
	}
	if len(audit.entries) != 1 || audit.entries[0].Result != base.AuditLogResultDenied {
		t.Fatalf("expected one denied entry, got %+v", audit.entries)
	}
}

// The whole point of the record is that a reveal cannot happen without one.
func TestRevealSecretsFailsClosedWhenNotRecordable(t *testing.T) {
	enableReveal(t, true)
	uc, _ := newRevealUC(t, true, errors.New("database is down"))
	setting := newBasicAuthSetting(t)

	err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)
	if err == nil {
		t.Fatal("a reveal that cannot be recorded must not happen")
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("the secret was revealed even though the record failed")
	}
}

// The config flag is the operator's, and it outranks the account's permissions -
// including an admin's, who passes every permission check.
func TestRevealSecretsRefusedWhenDisabledByConfig(t *testing.T) {
	enableReveal(t, false)
	uc, audit := newRevealUC(t, true, nil)
	setting := newBasicAuthSetting(t)

	err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)
	if !errors.Is(err, hperrors.ErrRevealSecretsDisabled) {
		t.Fatalf("expected the disabled-by-config error, got %v", err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("the secret must stay encrypted when the server disables reveal")
	}
	if len(audit.entries) != 1 || audit.entries[0].Result != base.AuditLogResultDenied {
		t.Error("a refusal by config must be recorded like any other")
	}
}

func TestRevealSecretsNotRequested(t *testing.T) {
	enableReveal(t, true)
	uc, audit := newRevealUC(t, true, nil)
	setting := newBasicAuthSetting(t)

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, false, setting); err != nil {
		t.Fatal(err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("nothing may be decrypted when reveal was not asked for")
	}
	if len(audit.entries) != 0 {
		t.Error("an ordinary read is not an audit event")
	}
}

// An inherited setting is read through this scope, not owned by it.
func TestRevealSecretsInheritedIsNeverRevealed(t *testing.T) {
	enableReveal(t, true)
	uc, audit := newRevealUC(t, true, nil)
	setting := newBasicAuthSetting(t)
	setting.CurrentObjectID = "obj_2"

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting); err != nil {
		t.Fatal(err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("an inherited setting must not hand out its secrets")
	}
	if len(audit.entries) != 0 {
		t.Error("nothing was revealed, so there is nothing to record")
	}
}

// A setting type with no secrets must not cost a permission check or a row.
func TestRevealSecretsOnTypeWithoutSecrets(t *testing.T) {
	enableReveal(t, true)
	uc, audit := newRevealUC(t, false, nil)

	setting := &entity.Setting{ID: "set_2", Name: "features", Type: base.SettingTypeAppFeatures}
	if err := setting.SetData(&entity.AppFeatureSettings{}); err != nil {
		t.Fatal(err)
	}
	setting.ObjectID, setting.CurrentObjectID = "obj_1", "obj_1"

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting); err != nil {
		t.Fatalf("a type without secrets must not be refused: %v", err)
	}
	if len(audit.entries) != 0 {
		t.Error("nothing was revealed, so there is nothing to record")
	}
}

// AuthorizeSecretReveal is the same gate the setting path uses, reached directly
// by callers whose secret is not a stored setting - the Swarm join token.
func TestAuthorizeSecretRevealRecordsSubjectAndDenial(t *testing.T) {
	enableReveal(t, true)
	uc, audit := newRevealUC(t, false, nil)

	subject := &RevealSubject{
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIGet,
		ResType: base.ResourceTypeCluster,
		Detail:  `{"tokenRole":"manager"}`,
	}
	err := uc.AuthorizeSecretReveal(context.Background(), nil, &basedto.Auth{}, subject)
	if err == nil {
		t.Fatal("want the refusal to come back")
	}

	if len(audit.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(audit.entries))
	}
	entry := audit.entries[0]
	// A refused attempt is the only sign of somebody trying doors, so it has to
	// be recorded rather than only refused.
	if entry.Result != base.AuditLogResultDenied {
		t.Errorf("result = %q", entry.Result)
	}
	if entry.Type != base.AuditLogTypeSecretReveal {
		t.Errorf("type = %q", entry.Type)
	}
	if entry.ResType != base.ResourceTypeCluster || entry.Detail != subject.Detail {
		t.Errorf("subject was not carried through: %q / %q", entry.ResType, entry.Detail)
	}
}

// The config flag is the operator's, and it refuses before the capability is even
// looked at - so an endpoint behind this gate stops working when it is off.
func TestAuthorizeSecretRevealRefusedWhenDisabledByConfig(t *testing.T) {
	enableReveal(t, false)
	uc, audit := newRevealUC(t, true, nil)

	err := uc.AuthorizeSecretReveal(context.Background(), nil, &basedto.Auth{}, &RevealSubject{
		Scope:   base.ObjectScopeGlobal,
		ResType: base.ResourceTypeCluster,
	})
	if !errors.Is(err, hperrors.ErrRevealSecretsDisabled) {
		t.Fatalf("want ErrRevealSecretsDisabled, got %v", err)
	}
	if len(audit.entries) != 1 || audit.entries[0].Result != base.AuditLogResultDenied {
		t.Error("the refusal must still be recorded")
	}
}
