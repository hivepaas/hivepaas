package sessionuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func TestPrivilegedSession(t *testing.T) {
	admin := &entity.User{ID: "admin-1", Username: "admin", Role: base.UserRoleAdmin}
	member := &entity.User{ID: "user-1", Username: "alice", Role: base.UserRoleMember}

	tests := []struct {
		name string
		req  *sessiondto.BaseCreateSessionReq
		want bool
	}{
		{
			name: "an admin signing in",
			req:  &sessiondto.BaseCreateSessionReq{User: admin, Method: auditMethodPassword},
			want: true,
		},
		{
			// The key is the long-lived credential; shortening the session it
			// mints guards nothing it does not already guard.
			name: "an admin's API key",
			req:  &sessiondto.BaseCreateSessionReq{User: admin, IsAPIKey: true, Method: auditMethodAPIKey},
			want: false,
		},
		{
			name: "a member signing in",
			req:  &sessiondto.BaseCreateSessionReq{User: member, Method: auditMethodPassword},
			want: false,
		},
		{
			name: "no user at all",
			req:  &sessiondto.BaseCreateSessionReq{Method: auditMethodPassword},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, privilegedSession(tt.req))
		})
	}
}

// A refresh mints a session the same way a login does, so the lifetimes follow
// the account as it is now rather than as it was when they first signed in.
func TestPrivilegedSessionFollowsTheCurrentRole(t *testing.T) {
	user := &entity.User{ID: "user-1", Role: base.UserRoleMember}
	req := &sessiondto.BaseCreateSessionReq{User: user, Method: auditMethodRefresh}

	assert.False(t, privilegedSession(req))

	user.Role = base.UserRoleAdmin
	assert.True(t, privilegedSession(req), "promoted, so the next renewal is the short one")
}

// The deadline has to hold where it matters: on the renewal that would otherwise
// carry the session past it. Checked before anything is minted, so a session
// that is over costs no token and no record.
func TestCreateSessionRefusesASessionPastItsDeadline(t *testing.T) {
	config.SetCurrent(&config.Config{})
	assert.NoError(t, jwtsession.InitJWTSession(&jwtsession.Config{
		Secret:          "test-signing-key",
		AccessTokenExp:  8 * time.Hour,
		RefreshTokenExp: 16 * time.Hour,
		SessionMaxExp:   24 * time.Hour,
	}))

	uc, audit := newAuditUCTest()

	_, err := uc.createSession(context.Background(), &sessiondto.BaseCreateSessionReq{
		User:      &entity.User{ID: "user-1", Role: base.UserRoleMember},
		StartedAt: timeutil.NowUTC().Add(-25 * time.Hour),
		Method:    auditMethodRefresh,
	})

	assert.ErrorIs(t, err, hperrors.ErrSessionJWTExpired)
	assert.Empty(t, audit.entries, "no session, so nothing to record")
}
