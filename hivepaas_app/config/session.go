package config

import "time"

type Session struct {
	LastAccessUpdatePeriod time.Duration `toml:"last_access_update_period" env:"HP_SESSION_LAST_ACCESS_UPDATE_PERIOD" default:"1m"` //nolint:lll
	PasscodeTimeout        time.Duration `toml:"passcode_timeout" env:"HP_SESSION_PASSCODE_TIMEOUT" default:"60s"`
	DeviceTrustedPeriod    time.Duration `toml:"device_trusted_period" env:"HP_SESSION_DEVICE_TRUSTED_PERIOD" default:"48h"` //nolint:lll

	JWTSecret       string        `toml:"jwt_secret" env:"HP_SESSION_JWT_SECRET"`
	AccessTokenExp  time.Duration `toml:"access_token_exp" env:"HP_SESSION_ACCESS_TOKEN_EXP" default:"8h"`
	RefreshTokenExp time.Duration `toml:"refresh_token_exp" env:"HP_SESSION_REFRESH_TOKEN_EXP" default:"16h"`

	// SessionMaxExp is how long a session may live from the login that started
	// it, however often it is renewed.
	//
	// The two lifetimes above are renewable, so on their own they cap how long a
	// session may sit unused rather than how long it may exist: a browser that
	// keeps making requests keeps getting fresh tokens for as long as anybody
	// leaves it open. This is the other half - a deadline set at the login and
	// carried in the session's own claims, which renewing cannot move.
	//
	// Zero means no deadline, which is what the install had before this existed.
	SessionMaxExp time.Duration `toml:"session_max_exp" env:"HP_SESSION_MAX_EXP" default:"24h"`

	// AdminAccessTokenExp and AdminRefreshTokenExp are the same two lifetimes for
	// a session signed into by an admin, which is the session that can change
	// anything in the install.
	//
	// The refresh lifetime is the one that decides how long such a session may
	// sit unused: the dashboard renews a session only when a request finds the
	// access token expired, and each renewal issues a fresh refresh token, so an
	// admin who keeps working is never signed out while an admin who walks away
	// is - after this long. The access lifetime is how long the token that
	// travels on every request stays usable if it is taken.
	//
	// Both are deliberately shorter than the shared pair rather than a separate
	// mechanism: an idle timeout that only some sessions have is the ordinary
	// control for privileged accounts, and it costs a working admin nothing.
	//
	// AdminSessionMaxExp is the deadline for those sessions - see SessionMaxExp.
	// Shorter than everybody else's for the same reason the other two are: a
	// session that can change the whole install is worth making somebody prove
	// who they are for again, sooner.
	//
	// Zero in any of the three means "no different from everybody else", so an
	// install that clears them is back to the shared lifetimes rather than to
	// something surprising.
	AdminAccessTokenExp  time.Duration `toml:"admin_access_token_exp" env:"HP_SESSION_ADMIN_ACCESS_TOKEN_EXP" default:"30m"`  //nolint:lll
	AdminRefreshTokenExp time.Duration `toml:"admin_refresh_token_exp" env:"HP_SESSION_ADMIN_REFRESH_TOKEN_EXP" default:"1h"` //nolint:lll
	AdminSessionMaxExp   time.Duration `toml:"admin_session_max_exp" env:"HP_SESSION_ADMIN_MAX_EXP" default:"8h"`

	BasicAuthUsername string `toml:"basic_auth_username" env:"HP_SESSION_BASIC_AUTH_USERNAME"`
	BasicAuthPassword string `toml:"basic_auth_password" env:"HP_SESSION_BASIC_AUTH_PASSWORD"`
}
