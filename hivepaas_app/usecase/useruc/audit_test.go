package useruc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
)

func newAuditUCTest() (*UC, *fakeAuditService) {
	audit := &fakeAuditService{}
	return &UC{auditService: audit}, audit
}

// targetUser is the account being acted on, actingAuth is somebody else doing it
// - the pair most of these entries are read for.
func targetUser() *entity.User {
	return &entity.User{ID: "user-1", Username: "alice", Role: base.UserRoleMember}
}

func actingAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User: &entity.User{ID: "admin-1", Username: "admin"},
	}}
}

func TestRecordUserChangeNamesBothSides(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordUserChange(context.Background(), nil, actingAuth(),
		base.AuditLogTypeUserUpdate, auditSectionAccount, targetUser(),
		auditdetail.New().Compare("role", base.UserRoleMember, base.UserRoleAdmin))

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, base.AuditLogTypeUserUpdate, entry.Type)
	assert.Equal(t, base.AuditLogSourceAPIUpdate, entry.Source)
	assert.Equal(t, auditSectionAccount, entry.Section)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	// The account acted on.
	assert.Equal(t, base.ObjectScopeUser, entry.Scope)
	assert.Equal(t, "user-1", entry.ObjectID)
	assert.Equal(t, "alice", entry.ResName)
	// The one who did it.
	assert.Equal(t, "admin-1", entry.Auth.User.ID)
	// A role is not a secret, and the values are the whole point of the entry.
	assert.Contains(t, entry.Detail, `"from":"member"`)
	assert.Contains(t, entry.Detail, `"to":"admin"`)
}

// An account appearing must not read as an edit, nor one being taken away.
func TestUserAuditSourceFollowsWhatHappened(t *testing.T) {
	tests := []struct {
		logType base.AuditLogType
		want    base.AuditLogSource
	}{
		{base.AuditLogTypeUserCreate, base.AuditLogSourceAPICreate},
		{base.AuditLogTypeUserDelete, base.AuditLogSourceAPIDelete},
		{base.AuditLogTypeUserUpdate, base.AuditLogSourceAPIUpdate},
	}

	for _, tt := range tests {
		t.Run(string(tt.logType), func(t *testing.T) {
			assert.Equal(t, tt.want, userAuditSource(tt.logType))
		})
	}
}

func TestRecordUserChangeRefusesWithNoAccount(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordUserChange(context.Background(), nil, actingAuth(),
		base.AuditLogTypeUserDelete, "", nil, nil)

	assert.Error(t, err)
	assert.Empty(t, audit.entries)
}

// The signup and reset links are possession of the address, which is what the
// account is - so the account is who the entry names.
func TestAuthOfUserAttributesTheAccountToItself(t *testing.T) {
	auth := authOfUser(targetUser())

	assert.Equal(t, "user-1", auth.User.ID)
	assert.Equal(t, "alice", auth.User.Username)
}

func TestRecordUserChangeDeniedKeepsTheAccountAndTheReason(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordUserChangeDenied(context.Background(), nil, actingAuth(),
		auditSectionMFARemove, targetUser(), auditReasonWrongPasscode)

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, base.AuditLogResultDenied, entry.Result)
	assert.Equal(t, auditSectionMFARemove, entry.Section)
	assert.Equal(t, "user-1", entry.ObjectID)
	assert.Contains(t, entry.Detail, `"reason":"wrong-passcode"`)
}

// The safety property, the same one the refused logins hold: an email box holds
// whatever was pasted into it, and none of it may reach an append-only table.
func TestRecordUserChangeDeniedNeverEchoesWhatWasTyped(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordUserChangeDenied(context.Background(), nil, nil,
		auditSectionPasswordResetRequest, nil, auditReasonUnknownEmail)

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, base.ObjectScopeGlobal, entry.Scope, "nothing to file it under")
	assert.Empty(t, entry.ObjectID)
	assert.Empty(t, entry.ResID)
	assert.Empty(t, entry.ResName)
	assert.Nil(t, entry.Auth)
	assert.Equal(t, `{"reason":"unknown-email"}`, entry.Detail)
}

// A link going out is not the account asking for it.
func TestRecordPasswordResetRequestNamesNoActor(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordPasswordResetRequest(context.Background(), nil, targetUser())

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Equal(t, auditSectionPasswordResetRequest, entry.Section)
	assert.Equal(t, "user-1", entry.ObjectID)
	assert.Nil(t, entry.Auth, "the form asks for an address and nothing else")
}

func TestForgotRefusalReason(t *testing.T) {
	tests := []struct {
		name string
		user *entity.User
		err  error
		want string
	}{
		{"no such address", nil, hperrors.ErrNotFound, auditReasonUnknownEmail},
		{"a real account the check refused", targetUser(), nil, auditReasonNotAllowed},
		// The install failing is not somebody being turned away.
		{"database unreachable", nil, errors.New("database is unreachable"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, forgotRefusalReason(tt.user, tt.err))
		})
	}
}

// Each form needs a section of its own, or the filter cannot tell a password
// change from a second factor being taken off.
func TestUserAuditSectionsAreDistinct(t *testing.T) {
	sections := []string{
		auditSectionInvite, auditSectionSignup, auditSectionAccount, auditSectionProfile,
		auditSectionPassword, auditSectionPasswordReset, auditSectionPasswordResetRequest,
		auditSectionMFASetup, auditSectionMFARemove,
	}

	seen := map[string]bool{}
	for _, section := range sections {
		assert.NotEmpty(t, section)
		assert.False(t, seen[section], "duplicate section: %s", section)
		seen[section] = true
	}
}

// An account change that could not be recorded does not happen.
func TestRecordUserChangeFailsWhenTheStoreIsDown(t *testing.T) {
	uc, audit := newAuditUCTest()
	audit.err = errors.New("audit store is down")

	err := uc.recordUserChange(context.Background(), nil, actingAuth(),
		base.AuditLogTypeUserUpdate, auditSectionPassword, targetUser(), nil)

	assert.Error(t, err)
}
