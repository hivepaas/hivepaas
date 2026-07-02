package jwtsession

import (
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var (
	testCfg = &Config{
		Secret:          "abc123",
		AccessTokenExp:  time.Second * 1,
		RefreshTokenExp: time.Second * 2,
	}
)

// initJWTSession resets sync.Once to be able to init JWT session multiple times
func initJWTSession(cfg *Config) error {
	once = sync.Once{}
	return InitJWTSession(cfg)
}

type testClaims struct {
	BaseClaims
	UserID string `json:"userId"`
}

func Test_GenerateToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &testClaims{UserID: "abc"}
		token, err := GenerateToken(claims, time.Second)
		assert.Nil(t, err)

		parsedClaims := &testClaims{}
		err = ParseToken(token, parsedClaims)
		assert.Nil(t, err)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
	})

	t.Run("failure with empty signing key", func(t *testing.T) {
		// Error when init with empty signing key
		newTestCfg := *testCfg
		newTestCfg.Secret = ""
		err := initJWTSession(&newTestCfg)
		assert.ErrorIs(t, err, ErrConfigInvalid)

		err = initJWTSession(testCfg)
		assert.Nil(t, err)

		// Error when generate with empty signing key
		claims := &testClaims{UserID: "abc"}
		signingKey = []byte{}
		_, err = GenerateToken(claims, time.Second)
		assert.ErrorIs(t, err, ErrConfigInvalid)
		assert.ErrorContains(t, err, "empty signing key")
	})
}

func Test_ParseToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &testClaims{UserID: "abc"}
		token, err := GenerateToken(claims, time.Second)
		assert.Nil(t, err)

		parsedClaims := &testClaims{}
		err = ParseToken(token, parsedClaims)
		assert.Nil(t, err)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)

		// Test token expiration
		time.Sleep(time.Second)
		parsedClaims = &testClaims{}
		err = ParseToken(token, parsedClaims)
		assert.ErrorIs(t, err, ErrTokenExpired)
	})

	t.Run("failure with changing signing method", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &testClaims{UserID: "abc"}
		// Use HS256 signing method
		signingMethod = jwt.SigningMethodHS256
		token, err := GenerateToken(claims, time.Second)
		assert.Nil(t, err)

		// Change the signing method then parse
		signingMethod = jwt.SigningMethodHS512
		err = ParseToken(token, &testClaims{})
		assert.ErrorIs(t, err, ErrTokenInvalid)
		assert.ErrorContains(t, err, "token is invalid")
	})

	t.Run("failure with changing signing key", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &testClaims{UserID: "abc"}
		signingKey = []byte("abc123")
		token, err := GenerateToken(claims, time.Second)
		assert.Nil(t, err)

		// Change the signing key then parse
		signingKey = []byte("abc124")
		err = ParseToken(token, &testClaims{})
		assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
	})

	t.Run("failure with empty signing key", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &testClaims{UserID: "abc"}
		signingKey = []byte("abc123")
		token, err := GenerateToken(claims, time.Second)
		assert.Nil(t, err)

		// Change the signing key then parse
		signingKey = []byte("")
		err = ParseToken(token, &testClaims{})
		assert.ErrorIs(t, err, ErrConfigInvalid)
		assert.ErrorContains(t, err, "empty signing key")
	})
}
