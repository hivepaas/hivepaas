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
	FuncNow         func() time.Time
}

var (
	once            sync.Once
	signingKey      []byte
	signingMethod   jwt.SigningMethod
	accessTokenExp  time.Duration
	refreshTokenExp time.Duration
	funcNow         func() time.Time
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
		funcNow = cfg.FuncNow
		if funcNow == nil {
			funcNow = timeutil.NowUTC
		}
	})
	return err
}
