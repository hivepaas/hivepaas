package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

/// Fakes

// fakePermissionManager stands in for the gate, which is tested where it lives -
// permissionimpl. What matters here is whether this layer asks at all, and what
// it says the secret is when it does.
type fakePermissionManager struct {
	permission.Manager
	err      error
	subjects []*permission.RevealSubject
}

func (f *fakePermissionManager) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	f.subjects = append(f.subjects, subject)
	return f.err
}

/// Helpers

func newRevealUC(refusal error) (*BaseUC, *fakePermissionManager) {
	perms := &fakePermissionManager{err: refusal}
	return &BaseUC{PermissionManager: perms}, perms
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

// revealScope is the scope the reveal is being asked for. It is what the audit
// entry is filed under, and an entry filed under the wrong scope is invisible to
// the scoped audit log listing - which is where somebody would go looking for it.
var revealScope = &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "obj_1"}

/// Tests

func TestRevealSecretsDecryptsWhenAuthorized(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newBasicAuthSetting(t)

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting); err != nil {
		t.Fatalf("reveal must succeed: %v", err)
	}

	if got := storedPassword(t, setting).String(); got != "the-real-password" {
		t.Errorf("secret was not revealed, got %q", got)
	}
	if len(perms.subjects) != 1 {
		t.Fatalf("want one authorization, got %d", len(perms.subjects))
	}
}

// The subject is what the audit entry is built from, so getting it wrong here
// files the record under something nobody will search for.
func TestRevealSecretsNamesTheSettingItIsAbout(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newBasicAuthSetting(t)

	_ = uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)

	subject := perms.subjects[0]
	if subject.Scope != base.ObjectScopeApp || subject.ObjectID != "obj_1" {
		t.Errorf("asked under %v/%q, want the scope the reveal was for", subject.Scope, subject.ObjectID)
	}
	if subject.ResType != base.ResourceTypeSetting || subject.ResID != "set_1" || subject.ResName != "web-auth" {
		t.Errorf("subject = %q/%q/%q", subject.ResType, subject.ResID, subject.ResName)
	}
	if subject.Source != base.AuditLogSourceAPIGet {
		t.Errorf("source = %q", subject.Source)
	}
	// A stored setting carries no secret type, which is what binds it to the
	// operator's flag rather than to any exemption.
	if subject.SecretType != "" {
		t.Errorf("a stored setting must have no secret type, got %q", subject.SecretType)
	}
}

func TestRevealSecretsLeavesTheSecretEncryptedWhenRefused(t *testing.T) {
	refusal := errors.New("not allowed")
	uc, _ := newRevealUC(refusal)
	setting := newBasicAuthSetting(t)

	err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)

	if !errors.Is(err, refusal) {
		t.Fatalf("the refusal must come back, got %v", err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("the secret must stay encrypted when the caller is refused")
	}
}

func TestRevealSecretsNotRequested(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newBasicAuthSetting(t)

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, false, setting); err != nil {
		t.Fatal(err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("nothing may be decrypted when reveal was not asked for")
	}
	if len(perms.subjects) != 0 {
		t.Error("an ordinary read must not cost an authorization or an audit row")
	}
}

// An inherited setting is read through this scope, not owned by it.
func TestRevealSecretsInheritedIsNeverRevealed(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newBasicAuthSetting(t)
	setting.CurrentObjectID = "obj_2"

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting); err != nil {
		t.Fatal(err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("an inherited setting must not hand out its secrets")
	}
	if len(perms.subjects) != 0 {
		t.Error("nothing was revealed, so there is nothing to authorize")
	}
}

// A setting type with no secrets must not cost a permission check or a row.
func TestRevealSecretsOnTypeWithoutSecrets(t *testing.T) {
	uc, perms := newRevealUC(errors.New("would refuse if asked"))

	setting := &entity.Setting{ID: "set_2", Name: "features", Type: base.SettingTypeAppFeatures}
	if err := setting.SetData(&entity.AppFeatureSettings{}); err != nil {
		t.Fatal(err)
	}
	setting.ObjectID, setting.CurrentObjectID = "obj_1", "obj_1"

	if err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting); err != nil {
		t.Fatalf("a type without secrets must not be refused: %v", err)
	}
	if len(perms.subjects) != 0 {
		t.Error("nothing was revealed, so there is nothing to authorize")
	}
}
