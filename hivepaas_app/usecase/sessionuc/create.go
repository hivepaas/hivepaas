package sessionuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	uidLen = 16
)

func (uc *UC) createSession(
	ctx context.Context,
	req *sessiondto.BaseCreateSessionReq,
) (resp *sessiondto.BaseCreateSessionResp, err error) {
	// Refused rather than recorded under an empty method. Every way into HivePaaS
	// ends here, so this is the one place that can make a login path added later
	// say how it authenticated - and a path that never says goes no further.
	if req.Method == "" {
		return nil, hperrors.NewArgumentInvalid("session login method")
	}

	// If user is demo user, restrict permissions to READ only
	if req.User.IsDemoUser() {
		if req.AccessAction == nil {
			req.AccessAction = &base.AccessActions{}
		}
		req.AccessAction.Reset(true, false, false, false)
	}

	authClaims := &jwtsession.AuthClaims{
		UID:          gofn.RandTokenAsHex(uidLen),
		UserID:       req.User.ID,
		IsAPIKey:     req.IsAPIKey,
		AccessAction: req.AccessAction,
	}

	// Decided per session rather than once at startup, because the answer depends
	// on who is signing in - and on a refresh it is re-decided from the account as
	// it is now, so an admin demoted mid-session lengthens on their next renewal
	// and a member promoted to admin shortens on theirs.
	exp := jwtsession.TokenExp(privilegedSession(req))

	// The deadline is measured from the login that began the session, so a
	// renewal cannot move it. Held to here rather than only checked, so the
	// tokens this hands out cannot outlive it either.
	startedAt := req.StartedAt
	if startedAt.IsZero() {
		startedAt = timeutil.NowUTC()
	}
	exp, alive := exp.Within(startedAt, timeutil.NowUTC())
	if !alive {
		return nil, hperrors.Wrap(hperrors.ErrSessionJWTExpired).
			WithMsgLog("session reached its %v deadline and has to be signed into again", exp.Max)
	}
	authClaims.StartedAt = startedAt.Unix()

	resp = &sessiondto.BaseCreateSessionResp{}
	resp.AccessToken, err = jwtsession.GenerateAccessTokenWithExp(authClaims, exp.Access)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create access token")
	}
	resp.AccessTokenExp = authClaims.ExpiresAt.Time

	resp.RefreshToken, err = jwtsession.GenerateRefreshTokenWithExp(authClaims, exp.Refresh)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create refresh token")
	}
	resp.RefreshTokenExp = authClaims.ExpiresAt.Time

	// Stores the uid in cache, so we can revoke the token later
	err = uc.userTokenRepo.Set(ctx, authClaims.UserID, authClaims.UID, resp.RefreshTokenExp.Sub(timeutil.NowUTC()))
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to store token in cache")
	}

	// Last, and the login does not happen if it cannot be recorded. Nothing has
	// left this function yet, so refusing here costs the caller another sign-in;
	// the alternative is a live session nobody can account for, which is the one
	// thing this record exists to prevent.
	if err = uc.recordLogin(ctx, req.User, req.Method); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

// privilegedSession reports whether this session gets the admin lifetimes.
//
// An admin account, signed into by a person. The role is the whole of the reason:
// an admin session can change anything the install has, so how long it stays
// usable unattended is worth deciding separately from everybody else's.
//
// Not a session minted from an API key, even an admin's. The key is itself the
// long-lived credential - it is issued for up to a year and is what an attacker
// would have to hold - so shortening the session it mints protects nothing it
// does not already protect, while multiplying how often every piece of
// automation has to sign in again. What guards a key is revoking it, which the
// api-key-revoke record next door is about.
func privilegedSession(req *sessiondto.BaseCreateSessionReq) bool {
	return !req.IsAPIKey && req.User != nil && req.User.IsAdmin()
}
