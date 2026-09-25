package permissionimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

/// Fakes

type fakeAuditService struct {
	auditservice.Service
	entries []*auditservice.Entry
	err     error
}

func (f *fakeAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return f.err
}

// fakeACLRepo grants nothing. Embedding the interface means only the one method
// this path reaches needs a body; anything else panics rather than passing.
type fakeACLRepo struct {
	repository.ACLPermissionRepo
}

func (f *fakeACLRepo) ListByResources(
	_ context.Context, _ database.IDB, _ []*base.PermissionResource, _ ...bunex.SelectQueryOption,
) ([]*entity.ACLPermission, error) {
	return nil, nil
}

/// Helpers

// newRevealManager builds the real manager, because the gates are what is being
// tested. A fake manager here would only ever test itself.
func newRevealManager(auditErr error) (*manager, *fakeAuditService) {
	audit := &fakeAuditService{err: auditErr}
	return &manager{aclPermissionRepo: &fakeACLRepo{}, auditService: audit}, audit
}

// adminAuth passes the capability gate. CheckAccess answers admins yes without
// consulting anything, which is exactly why a denial has to be recorded rather
// than merely returned - for an admin there is no denial to see.
func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "usr_admin", Role: base.UserRoleAdmin}}}
}

// plainAuth holds no capability, so the ACL lookup decides - and finds nothing.
func plainAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "usr_1", Role: base.UserRoleMember}}}
}

func enableReveal(t *testing.T, enabled bool, alwaysReturn ...string) {
	t.Helper()
	prev := config.Current()
	config.SetCurrent(&config.Config{Security: config.Security{
		ReturnSecretsViaAPI:     enabled,
		AlwaysReturnSecretTypes: alwaysReturn,
	}})
	t.Cleanup(func() { config.SetCurrent(prev) })
}

func joinTokenSubject() *permission.RevealSubject {
	return &permission.RevealSubject{
		Scope:    base.ObjectScopeGlobal,
		ObjectID: "obj_1",
		Source:   base.AuditLogSourceAPIGet,
		ResType:  base.ResourceTypeCluster,
		ResID:    "cluster",
		ResName:  "swarm",
		Detail:   `{"tokenRole":"manager"}`,
	}
}

/// Tests

func TestAuthorizeSecretRevealAllowsAndRecords(t *testing.T) {
	enableReveal(t, true)
	mgr, audit := newRevealManager(nil)

	if err := mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), joinTokenSubject()); err != nil {
		t.Fatalf("an admin with the flag on must be allowed: %v", err)
	}

	if len(audit.entries) != 1 {
		t.Fatalf("want one entry, got %d", len(audit.entries))
	}
	if audit.entries[0].Result != base.AuditLogResultAllowed {
		t.Errorf("result = %q", audit.entries[0].Result)
	}
}

// The refusal is the half that matters: an admin passes the capability gate
// unconditionally, so the attempt is the only thing there is to see.
func TestAuthorizeSecretRevealRecordsADenial(t *testing.T) {
	enableReveal(t, true)
	mgr, audit := newRevealManager(nil)

	err := mgr.AuthorizeSecretReveal(context.Background(), nil, plainAuth(), joinTokenSubject())

	if !errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets) {
		t.Fatalf("want the missing-capability error, got %v", err)
	}
	if len(audit.entries) != 1 || audit.entries[0].Result != base.AuditLogResultDenied {
		t.Fatalf("a refusal must be recorded, got %+v", audit.entries)
	}
}

// Everything the entry is filed under has to survive the trip, or it lands
// somewhere nobody will look for it.
func TestAuthorizeSecretRevealCarriesTheSubjectIntoTheEntry(t *testing.T) {
	enableReveal(t, true)
	mgr, audit := newRevealManager(nil)
	subject := joinTokenSubject()

	_ = mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), subject)

	entry := audit.entries[0]
	if entry.Type != base.AuditLogTypeSecretReveal {
		t.Errorf("type = %q", entry.Type)
	}
	if entry.Scope != subject.Scope || entry.ObjectID != subject.ObjectID {
		t.Errorf("filed under %v/%q, want %v/%q", entry.Scope, entry.ObjectID, subject.Scope, subject.ObjectID)
	}
	if entry.Source != subject.Source {
		t.Errorf("source = %q", entry.Source)
	}
	if entry.ResType != subject.ResType || entry.ResID != subject.ResID || entry.ResName != subject.ResName {
		t.Errorf("resource = %q/%q/%q", entry.ResType, entry.ResID, entry.ResName)
	}
	if entry.Detail != subject.Detail {
		t.Errorf("detail = %q", entry.Detail)
	}
}

// A secret released without a trace is what this path exists to prevent, so a
// reveal that cannot be recorded does not happen.
func TestAuthorizeSecretRevealFailsClosedWhenNotRecordable(t *testing.T) {
	enableReveal(t, true)
	mgr, _ := newRevealManager(errors.New("database is down"))

	err := mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), joinTokenSubject())

	if err == nil {
		t.Fatal("a reveal that cannot be recorded must be refused")
	}
}

// The config flag is the operator's and outranks the account's permissions,
// including an admin's.
func TestAuthorizeSecretRevealRefusedWhenDisabledByConfig(t *testing.T) {
	enableReveal(t, false)
	mgr, audit := newRevealManager(nil)

	err := mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), joinTokenSubject())

	if !errors.Is(err, hperrors.ErrRevealSecretsDisabled) {
		t.Fatalf("want the disabled-by-config error, got %v", err)
	}
	if len(audit.entries) != 1 || audit.entries[0].Result != base.AuditLogResultDenied {
		t.Error("a refusal by config is recorded like any other")
	}
}

// An operator can stand the flag down for one named kind of secret - which is how
// node onboarding keeps working without opening up every stored credential.
func TestAuthorizeSecretRevealHonoursASecretTypeExemption(t *testing.T) {
	enableReveal(t, false, string(base.SecretTypeSwarmJoinToken))
	mgr, audit := newRevealManager(nil)

	subject := joinTokenSubject()
	subject.SecretType = base.SecretTypeSwarmJoinToken

	if err := mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), subject); err != nil {
		t.Fatalf("an exempted secret type must pass the flag: %v", err)
	}
	if audit.entries[0].Result != base.AuditLogResultAllowed {
		t.Error("the exemption should have produced an allowed entry")
	}
}

// The exemption stands the operator's flag down, not the account's capability. A
// caller who may not reveal secrets still may not.
func TestAuthorizeSecretRevealExemptionDoesNotWaiveTheCapability(t *testing.T) {
	enableReveal(t, false, string(base.SecretTypeSwarmJoinToken))
	mgr, audit := newRevealManager(nil)

	subject := joinTokenSubject()
	subject.SecretType = base.SecretTypeSwarmJoinToken

	err := mgr.AuthorizeSecretReveal(context.Background(), nil, plainAuth(), subject)

	if !errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets) {
		t.Fatalf("want the missing-capability error, got %v", err)
	}
	if audit.entries[0].Result != base.AuditLogResultDenied {
		t.Error("the refusal must still be recorded")
	}
}

// A stored setting carries no secret type, and is bound by the flag whatever the
// exemption list says.
func TestAuthorizeSecretRevealBindsAnUntypedSecretToTheFlag(t *testing.T) {
	enableReveal(t, false, string(base.SecretTypeSwarmJoinToken))
	mgr, _ := newRevealManager(nil)

	err := mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), joinTokenSubject())

	if !errors.Is(err, hperrors.ErrRevealSecretsDisabled) {
		t.Fatalf("want the disabled-by-config error, got %v", err)
	}
}

func TestAuthorizeSecretRevealRefusesAMissingSubject(t *testing.T) {
	enableReveal(t, true)
	mgr, audit := newRevealManager(nil)

	if err := mgr.AuthorizeSecretReveal(context.Background(), nil, adminAuth(), nil); err == nil {
		t.Fatal("a reveal with no subject has nothing to record and must be refused")
	}
	if len(audit.entries) != 0 {
		t.Error("nothing was decided, so there is nothing to record")
	}
}

// Import's validate asks without recording: nothing is revealed until apply.
func TestMayRevealSecretsAnswersWithoutRecording(t *testing.T) {
	enableReveal(t, true)
	mgr, audit := newRevealManager(nil)

	allowed, err := mgr.MayRevealSecrets(context.Background(), nil, adminAuth())
	assert.NoError(t, err)
	assert.True(t, allowed)
	allowed, err = mgr.MayRevealSecrets(context.Background(), nil, plainAuth())
	assert.NoError(t, err)
	assert.False(t, allowed, "a denial is an answer, not an error")

	enableReveal(t, false)
	allowed, err = mgr.MayRevealSecrets(context.Background(), nil, adminAuth())
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Empty(t, audit.entries)
}
