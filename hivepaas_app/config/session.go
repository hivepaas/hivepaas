package config

import "time"

type Session struct {
	LastAccessUpdatePeriod time.Duration `toml:"last_access_update_period" env:"HP_SESSION_LAST_ACCESS_UPDATE_PERIOD" default:"1m"` //nolint:lll
	PasscodeTimeout        time.Duration `toml:"passcode_timeout" env:"HP_SESSION_PASSCODE_TIMEOUT" default:"60s"`
	DeviceTrustedPeriod    time.Duration `toml:"device_trusted_period" env:"HP_SESSION_DEVICE_TRUSTED_PERIOD" default:"48h"` //nolint:lll

	JWTSecret       string        `toml:"jwt_secret" env:"HP_SESSION_JWT_SECRET"`
	AccessTokenExp  time.Duration `toml:"access_token_exp" env:"HP_SESSION_ACCESS_TOKEN_EXP" default:"12h"`
	RefreshTokenExp time.Duration `toml:"refresh_token_exp" env:"HP_SESSION_REFRESH_TOKEN_EXP" default:"24h"`

	BasicAuthUsername string `toml:"basic_auth_username" env:"HP_SESSION_BASIC_AUTH_USERNAME"`
	BasicAuthPassword string `toml:"basic_auth_password" env:"HP_SESSION_BASIC_AUTH_PASSWORD"`
}
