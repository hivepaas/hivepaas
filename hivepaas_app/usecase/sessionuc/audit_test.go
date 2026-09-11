package sessionuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// spyAuditService keeps what was recorded, so a test can assert on the record
// itself rather than only on the answer the caller got.
type spyAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (s *spyAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	s.entries = append(s.entries, entry)
	return s.err
}

func newAuditUCTest() (*UC, *spyAuditService) {
	audit := &spyAuditService{}
	return &UC{auditService: audit}, audit
}

func namedUser() *entity.User {
	return &entity.User{ID: "user-1", Username: "alice"}
}

func TestRecordLoginNamesTheUserAndTheWayIn(t *testing.T) {
	uc, audit := newAuditUCTest()

	assert.NoError(t, uc.recordLogin(context.Background(), namedUser(), auditMethodPassword))

	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, base.AuditLogTypeUserLogin, entry.Type)
	assert.Equal(t, base.ObjectScopeUser, entry.Scope)
	assert.Equal(t, "user-1", entry.ObjectID)
	assert.Equal(t, auditMethodPassword, entry.Section)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Equal(t, base.ResourceTypeUser, entry.ResType)
	assert.Equal(t, "alice", entry.ResName)
	// A login is the one act where the actor is established by the request itself.
	assert.Equal(t, "user-1", entry.Auth.User.ID)
}

// A session renewing itself is not somebody signing in, and recording every
// renewal would bury the logins under them.
func TestRecordLoginSkipsARefresh(t *testing.T) {
	uc, audit := newAuditUCTest()

	assert.NoError(t, uc.recordLogin(context.Background(), namedUser(), auditMethodRefresh))

	assert.Empty(t, audit.entries)
}

func TestRecordLoginDeniedNamesTheAccountButNoActor(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordLoginDenied(context.Background(), namedUser(),
		auditMethodPassword, auditReasonWrongPassword)

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, base.AuditLogResultDenied, entry.Result)
	assert.Equal(t, base.ObjectScopeUser, entry.Scope)
	assert.Equal(t, "user-1", entry.ObjectID)
	assert.Equal(t, "alice", entry.ResName, "the account that was aimed at")
	assert.Contains(t, entry.Detail, `"reason":"wrong-password"`)
	// Who was on the other end is what the request failed to establish.
	assert.Nil(t, entry.Auth)
}

// The safety property: a username box holds a password often enough that
// echoing what was typed would turn this table into a place to find them.
func TestRecordLoginDeniedNeverEchoesWhatWasTyped(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordLoginDenied(context.Background(), nil,
		auditMethodPassword, auditReasonUnknownUser)

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	// Nothing to file it under, so it goes to the install's own log.
	assert.Equal(t, base.ObjectScopeGlobal, entry.Scope)
	assert.Empty(t, entry.ObjectID)
	assert.Empty(t, entry.ResID)
	assert.Empty(t, entry.ResName)
	assert.Equal(t, `{"reason":"unknown-user"}`, entry.Detail,
		"the reason is the whole of what a refused unknown identifier may say")
}

func TestRecordLogoutTellsTheTwoKindsApart(t *testing.T) {
	for _, section := range []string{auditSectionLogoutCurrent, auditSectionLogoutAll} {
		uc, audit := newAuditUCTest()
		user := &basedto.User{
			User:       namedUser(),
			AuthClaims: &jwtsession.AuthClaims{UserID: "user-1", UID: "session-1"},
		}

		assert.NoError(t, uc.recordLogout(context.Background(), user, section))

		assert.Len(t, audit.entries, 1)
		entry := audit.entries[0]
		assert.Equal(t, base.AuditLogTypeUserLogout, entry.Type)
		assert.Equal(t, base.AuditLogSourceAPIDelete, entry.Source)
		assert.Equal(t, section, entry.Section)
		assert.Equal(t, "user-1", entry.ObjectID)
	}
}

// The id has to come from the token being given up, not from a row that may not
// have been read.
func TestRecordLogoutRefusesASessionWithNoClaims(t *testing.T) {
	uc, audit := newAuditUCTest()

	err := uc.recordLogout(context.Background(), &basedto.User{User: namedUser()},
		auditSectionLogoutCurrent)

	assert.Error(t, err)
	assert.Empty(t, audit.entries)
}

// A login path that does not say how it authenticated goes no further - which is
// what stops one added later from being the door with no record.
func TestCreateSessionRefusesAnUnnamedMethod(t *testing.T) {
	uc, audit := newAuditUCTest()

	_, err := uc.createSession(context.Background(), &sessiondto.BaseCreateSessionReq{
		User: namedUser(),
	})

	assert.Error(t, err)
	assert.Empty(t, audit.entries, "nothing is recorded for a session that was not created")
}

func TestLoginRefusalReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"wrong password", hperrors.ErrPasswordMismatched, auditReasonWrongPassword},
		{"locked out", hperrors.ErrTooManyLoginFailures, auditReasonLockedOut},
		{"sso required", hperrors.ErrSSORequired, auditReasonSSORequired},
		// The install failing is not somebody being turned away, and recording it
		// as one would put bad days in among the attempts.
		{"redis unreachable", errors.New("redis is unreachable"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, loginRefusalReason(hperrors.Wrap(tt.err)))
		})
	}
}

func TestAPIKeyRefusalReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"no such key", hperrors.ErrNotFound, auditReasonInvalidAPIKey},
		{"key switched off", hperrors.ErrAPIKeyInvalid, auditReasonInvalidAPIKey},
		{"secret does not match", hperrors.ErrMismatch, auditReasonInvalidAPIKey},
		{"database unreachable", errors.New("database is unreachable"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, apiKeyRefusalReason(hperrors.Wrap(tt.err)))
		})
	}
}

// An action that could not be recorded does not happen: the caller signs in
// again, which is cheaper than a live session nobody can account for.
func TestLoginFailsWhenItCannotBeRecorded(t *testing.T) {
	uc, audit := newAuditUCTest()
	audit.err = errors.New("audit store is down")

	err := uc.recordLogin(context.Background(), namedUser(), auditMethodPassword)

	assert.Error(t, err)
}
