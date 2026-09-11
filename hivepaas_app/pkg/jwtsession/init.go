package jwtsession

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

var (
	ErrConfigInvalid = errors.New("configuration is invalid")
	ErrTokenInvalid  = errors.New("token is invalid")
	ErrTokenExpired  = errors.New("token expired")
)

type Config struct {
	Secret          string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration

	// SessionMaxExp is how long a session may live from its original login,
	// however often it is renewed. Zero means no deadline.
	SessionMaxExp time.Duration

	// PrivilegedAccessTokenExp, PrivilegedRefreshTokenExp and
	// PrivilegedSessionMaxExp are the same three when the caller asks for them -
	// see TokenExp. Zero means the shared one, which is what an install that has
	// not asked for a distinction gets.
	PrivilegedAccessTokenExp  time.Duration
	PrivilegedRefreshTokenExp time.Duration
	PrivilegedSessionMaxExp   time.Duration

	FuncNow func() time.Time
}

var (
	once                      sync.Once
	signingKey                []byte
	signingMethod             jwt.SigningMethod
	accessTokenExp            time.Duration
	refreshTokenExp           time.Duration
	sessionMaxExp             time.Duration
	privilegedAccessTokenExp  time.Duration
	privilegedRefreshTokenExp time.Duration
	privilegedSessionMaxExp   time.Duration
	funcNow                   func() time.Time
)

// InitJWTSession initializes variables for JWT Session
func InitJWTSession(cfg *Config) (err error) {
	once.Do(func() {
		signingKey = []byte(cfg.Secret)
		// NOTE: jwt allows empty secret key, we should report the error here
		if len(signingKey) == 0 {
			err = fmt.Errorf("empty signing key: %w", ErrConfigInvalid)
			return
		}
		signingMethod = jwt.SigningMethodHS256
		accessTokenExp = cfg.AccessTokenExp
		refreshTokenExp = cfg.RefreshTokenExp
		sessionMaxExp = cfg.SessionMaxExp
		privilegedAccessTokenExp = cfg.PrivilegedAccessTokenExp
		privilegedRefreshTokenExp = cfg.PrivilegedRefreshTokenExp
		privilegedSessionMaxExp = cfg.PrivilegedSessionMaxExp
		funcNow = cfg.FuncNow
		if funcNow == nil {
			funcNow = timeutil.NowUTC
		}
	})
	return err
}
