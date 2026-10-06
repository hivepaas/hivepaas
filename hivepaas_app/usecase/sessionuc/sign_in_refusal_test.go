package sessionuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// unavailableUsers are the accounts no way in may open a session for.
func unavailableUsers() map[string]*entity.User {
	past := timeutil.NowUTC().Add(-time.Hour)
	return map[string]*entity.User{
		"disabled": {ID: "u-disabled", Username: "dora", Status: base.UserStatusDisabled},
		"pending":  {ID: "u-pending", Username: "pete", Status: base.UserStatusPending},
		"expired":  {ID: "u-expired", Username: "eve", Status: base.UserStatusActive, AccessExpireAt: past},
	}
}

// Every way in ends in createSession: a disabled, pending or expired account
// gets no session there, whichever way it came - password, passcode, SSO, API
// key, or a refresh of a session it had before.
func TestCreateSessionRefusesAUserWhoCannotSignIn(t *testing.T) {
	for name, user := range unavailableUsers() {
		t.Run(name, func(t *testing.T) {
			uc, audit := newAuditUCTest()

			_, err := uc.createSession(context.Background(), &sessiondto.BaseCreateSessionReq{
				User: user, Method: auditMethodPassword,
			})

			assert.ErrorIs(t, err, hperrors.ErrUserUnavailable)
			assert.Empty(t, audit.entries, "no login recorded")
		})
	}
}

type userByName struct {
	repository.UserRepo
	user *entity.User
}

func (r *userByName) GetByUsernameOrEmail(context.Context, database.IDB, string, string,
	...bunex.SelectQueryOption) (*entity.User, error) {
	return r.user, nil
}

type rightPassword struct{ userservice.Service }

func (rightPassword) VerifyPassword(*entity.User, string) error { return nil }

// The right password of a disabled account is refused, before any second
// factor is asked for, and the refusal is recorded as one.
func TestLoginWithPasswordRefusesADisabledUser(t *testing.T) {
	uc, audit := newAuditUCTest()
	user := unavailableUsers()["disabled"]
	user.TotpSecret = "a second factor"
	uc.userRepo = &userByName{user: user}
	uc.userService = rightPassword{}
	uc.cacheLoginAttemptRepo = &fakeLoginAttemptRepo{}

	resp, err := uc.LoginWithPassword(context.Background(),
		&sessiondto.LoginWithPasswordReq{Username: "dora", Password: "right"})

	assert.ErrorIs(t, err, hperrors.ErrUserUnavailable)
	assert.Nil(t, resp, "no session, and no second factor asked for")
	if assert.Len(t, audit.entries, 1) {
		assert.Equal(t, base.AuditLogResultDenied, audit.entries[0].Result)
		assert.Contains(t, audit.entries[0].Detail, auditReasonUserUnavailable)
	}
}
