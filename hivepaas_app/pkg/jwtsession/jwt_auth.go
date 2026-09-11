package jwtsession

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// AuthClaims the claims for authentication
type AuthClaims struct {
	BaseClaims
	UID       string `json:"uid"`
	UserID    string `json:"userId"`
	IsRefresh bool   `json:"isRefresh,omitempty"`
	IsAPIKey  bool   `json:"isAPIKey,omitempty"`

	// StartedAt is when the login that began this session happened, as unix
	// seconds, carried through every renewal so the deadline is measured from the
	// login rather than from the last renewal. Separate from the registered
	// IssuedAt, which is per token and moves every time one is minted.
	//
	// Zero means the session predates this claim - a session already live when an
	// install upgraded. Those are let through and dated from their next renewal
	// rather than from the epoch, which would sign everybody out at the deploy.
	StartedAt int64 `json:"startedAt,omitempty"`

	AccessAction *base.AccessActions `json:"access,omitempty"`
}

// SessionStartedAt is when this session's login happened, or the zero time when
// the session predates the claim.
func (c *AuthClaims) SessionStartedAt() time.Time {
	if c == nil || c.StartedAt <= 0 {
		return time.Time{}
	}
	return time.Unix(c.StartedAt, 0).UTC()
}

// SessionExp is how long a session and its tokens last.
//
// Access and Refresh are renewable and say how long a session may go unused:
// each renewal issues a fresh pair, so a browser that keeps making requests
// keeps the session alive. Max is the part renewing cannot move - the deadline
// set at the login itself. Zero Max means no deadline.
type SessionExp struct {
	Access  time.Duration
	Refresh time.Duration
	Max     time.Duration
}

// TokenExp is what a new session gets.
//
// privileged asks for the set an install keeps for the sessions that can change
// it. Any of those being unset answers with the shared value, so an install that
// has not asked for the distinction does not get part of one.
//
// The numbers are the operator's and are not second-guessed here: an install
// that wants its privileged sessions to last longer than everybody else's is
// making a decision, not a mistake this function should quietly correct.
func TokenExp(privileged bool) SessionExp {
	exp := SessionExp{Access: accessTokenExp, Refresh: refreshTokenExp, Max: sessionMaxExp}
	if !privileged {
		return exp
	}
	if privilegedAccessTokenExp > 0 {
		exp.Access = privilegedAccessTokenExp
	}
	if privilegedRefreshTokenExp > 0 {
		exp.Refresh = privilegedRefreshTokenExp
	}
	if privilegedSessionMaxExp > 0 {
		exp.Max = privilegedSessionMaxExp
	}
	return exp
}

// Within narrows the token lifetimes so neither outlives the session's deadline,
// and reports whether the session may go on at all.
//
// Clamping rather than only refusing at the deadline, because a token that
// outlived the session would be exactly the hole the deadline is there to close:
// the last renewal before a deadline would otherwise hand out a token good for
// hours past it, and nothing would come back to check.
//
// A session with no deadline, or one whose start is not known, is returned
// unchanged and alive - see AuthClaims.StartedAt for why an unknown start is not
// treated as the beginning of time.
func (e SessionExp) Within(startedAt, now time.Time) (SessionExp, bool) {
	if e.Max <= 0 || startedAt.IsZero() {
		return e, true
	}

	remaining := startedAt.Add(e.Max).Sub(now)
	if remaining <= 0 {
		return e, false
	}
	if e.Access > remaining {
		e.Access = remaining
	}
	if e.Refresh > remaining {
		e.Refresh = remaining
	}
	return e, true
}

// GenerateAccessToken generates access token with the shared lifetime.
func GenerateAccessToken(authClaims *AuthClaims) (string, error) {
	return GenerateAccessTokenWithExp(authClaims, accessTokenExp)
}

// GenerateAccessTokenWithExp generates access token lasting exp.
//
// For a caller that decides the lifetime per session rather than taking the
// shared one - see TokenExp.
func GenerateAccessTokenWithExp(authClaims *AuthClaims, exp time.Duration) (string, error) {
	authClaims.IsRefresh = false
	token, err := GenerateToken(authClaims, exp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

// GenerateRefreshToken generates refresh token with the shared lifetime.
func GenerateRefreshToken(authClaims *AuthClaims) (string, error) {
	return GenerateRefreshTokenWithExp(authClaims, refreshTokenExp)
}

// GenerateRefreshTokenWithExp generates refresh token lasting exp.
func GenerateRefreshTokenWithExp(authClaims *AuthClaims, exp time.Duration) (string, error) {
	authClaims.IsRefresh = true
	token, err := GenerateToken(authClaims, exp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}
