package jwtsession

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateAccessToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &AuthClaims{UserID: "abc"}
		token, err := GenerateAccessToken(claims)
		assert.Nil(t, err)

		parsedClaims := &AuthClaims{}
		err = ParseToken(token, parsedClaims)
		assert.Nil(t, err)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
	})

	t.Run("failure", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &AuthClaims{UserID: "abc"}
		signingKey = []byte("")
		_, err := GenerateAccessToken(claims)
		assert.ErrorIs(t, err, ErrConfigInvalid)
		assert.ErrorContains(t, err, "empty signing key")
	})
}

func Test_GenerateRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &AuthClaims{UserID: "abc"}
		token, err := GenerateRefreshToken(claims)
		assert.Nil(t, err)

		parsedClaims := &AuthClaims{}
		err = ParseToken(token, parsedClaims)
		assert.Nil(t, err)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
	})

	t.Run("failure", func(t *testing.T) {
		_ = initJWTSession(testCfg)

		claims := &AuthClaims{UserID: "abc"}
		signingKey = []byte("")
		_, err := GenerateRefreshToken(claims)
		assert.ErrorIs(t, err, ErrConfigInvalid)
		assert.ErrorContains(t, err, "empty signing key")
	})
}
