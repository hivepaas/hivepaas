package jwtsession

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	sharedOnlyCfg = &Config{
		Secret:          "abc123",
		AccessTokenExp:  8 * time.Hour,
		RefreshTokenExp: 16 * time.Hour,
		SessionMaxExp:   24 * time.Hour,
	}
	privilegedCfg = &Config{
		Secret:                    "abc123",
		AccessTokenExp:            8 * time.Hour,
		RefreshTokenExp:           16 * time.Hour,
		SessionMaxExp:             24 * time.Hour,
		PrivilegedAccessTokenExp:  30 * time.Minute,
		PrivilegedRefreshTokenExp: time.Hour,
		PrivilegedSessionMaxExp:   8 * time.Hour,
	}
)

func TestTokenExp(t *testing.T) {
	t.Run("an ordinary session takes the shared set", func(t *testing.T) {
		assert.NoError(t, initJWTSession(privilegedCfg))

		exp := TokenExp(false)
		assert.Equal(t, 8*time.Hour, exp.Access)
		assert.Equal(t, 16*time.Hour, exp.Refresh)
		assert.Equal(t, 24*time.Hour, exp.Max)
	})

	t.Run("a privileged session takes the shorter set", func(t *testing.T) {
		assert.NoError(t, initJWTSession(privilegedCfg))

		exp := TokenExp(true)
		assert.Equal(t, 30*time.Minute, exp.Access)
		assert.Equal(t, time.Hour, exp.Refresh)
		assert.Equal(t, 8*time.Hour, exp.Max)
	})

	// An install that never set them must not end up with sessions that expire at
	// once, which is what a zero would mean if it were taken as a lifetime.
	t.Run("unset privileged values fall back to the shared ones", func(t *testing.T) {
		assert.NoError(t, initJWTSession(sharedOnlyCfg))

		exp := TokenExp(true)
		assert.Equal(t, 8*time.Hour, exp.Access)
		assert.Equal(t, 16*time.Hour, exp.Refresh)
		assert.Equal(t, 24*time.Hour, exp.Max)
	})

	t.Run("only part of the set configured", func(t *testing.T) {
		assert.NoError(t, initJWTSession(&Config{
			Secret:                  "abc123",
			AccessTokenExp:          8 * time.Hour,
			RefreshTokenExp:         16 * time.Hour,
			SessionMaxExp:           24 * time.Hour,
			PrivilegedSessionMaxExp: 8 * time.Hour,
		}))

		exp := TokenExp(true)
		assert.Equal(t, 8*time.Hour, exp.Access, "not asked for, so unchanged")
		assert.Equal(t, 16*time.Hour, exp.Refresh)
		assert.Equal(t, 8*time.Hour, exp.Max)
	})
}

func TestSessionExpWithin(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	exp := SessionExp{Access: 30 * time.Minute, Refresh: time.Hour, Max: 8 * time.Hour}

	t.Run("early in the session nothing is narrowed", func(t *testing.T) {
		got, alive := exp.Within(now.Add(-time.Hour), now)

		assert.True(t, alive)
		assert.Equal(t, 30*time.Minute, got.Access)
		assert.Equal(t, time.Hour, got.Refresh)
	})

	// The renewal that would otherwise hand out a token good for hours past the
	// deadline - the hole the deadline exists to close.
	t.Run("near the deadline both are cut to what is left", func(t *testing.T) {
		got, alive := exp.Within(now.Add(-7*time.Hour-50*time.Minute), now)

		assert.True(t, alive)
		assert.Equal(t, 10*time.Minute, got.Access)
		assert.Equal(t, 10*time.Minute, got.Refresh)
	})

	t.Run("past the deadline the session is over", func(t *testing.T) {
		_, alive := exp.Within(now.Add(-8*time.Hour-time.Second), now)

		assert.False(t, alive)
	})

	t.Run("exactly at the deadline the session is over", func(t *testing.T) {
		_, alive := exp.Within(now.Add(-8*time.Hour), now)

		assert.False(t, alive)
	})

	t.Run("no deadline configured leaves everything alone", func(t *testing.T) {
		noMax := SessionExp{Access: 8 * time.Hour, Refresh: 16 * time.Hour}

		got, alive := noMax.Within(now.Add(-100*time.Hour), now)

		assert.True(t, alive)
		assert.Equal(t, noMax, got)
	})

	// A session that predates the claim: dated from its next renewal rather than
	// from the epoch, which would sign everybody out at the deploy that adds it.
	t.Run("an unknown start is not the beginning of time", func(t *testing.T) {
		got, alive := exp.Within(time.Time{}, now)

		assert.True(t, alive)
		assert.Equal(t, exp, got)
	})
}

func TestSessionStartedAt(t *testing.T) {
	started := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

	claims := &AuthClaims{StartedAt: started.Unix()}
	assert.Equal(t, started, claims.SessionStartedAt())

	assert.True(t, (&AuthClaims{}).SessionStartedAt().IsZero(), "a session from before the claim")
	assert.True(t, (*AuthClaims)(nil).SessionStartedAt().IsZero())
}

// The lifetime the caller asks for is the one the token gets, which is what lets
// two sessions minted by the same code expire at different times.
func TestGenerateTokenWithExpHonoursTheLifetime(t *testing.T) {
	assert.NoError(t, initJWTSession(sharedOnlyCfg))

	claims := &AuthClaims{UserID: "user-1"}
	_, err := GenerateAccessTokenWithExp(claims, 30*time.Minute)
	assert.NoError(t, err)
	assert.WithinDuration(t, funcNow().Add(30*time.Minute), claims.ExpiresAt.Time, time.Minute)

	_, err = GenerateRefreshTokenWithExp(claims, time.Hour)
	assert.NoError(t, err)
	assert.WithinDuration(t, funcNow().Add(time.Hour), claims.ExpiresAt.Time, time.Minute)
	assert.True(t, claims.IsRefresh)
}

// The start has to survive the round trip, or every renewal would look like a
// fresh login and the deadline would never arrive.
func TestStartedAtSurvivesTheToken(t *testing.T) {
	assert.NoError(t, initJWTSession(sharedOnlyCfg))
	started := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Second)

	claims := &AuthClaims{UserID: "user-1", StartedAt: started.Unix()}
	token, err := GenerateRefreshTokenWithExp(claims, time.Hour)
	assert.NoError(t, err)

	parsed := &AuthClaims{}
	assert.NoError(t, ParseToken(token, parsed))
	assert.Equal(t, started, parsed.SessionStartedAt())
}
