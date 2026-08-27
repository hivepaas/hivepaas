package jwtsession

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// AuthClaims the claims for authentication
type AuthClaims struct {
	BaseClaims
	UID          string              `json:"uid"`
	UserID       string              `json:"userId"`
	IsRefresh    bool                `json:"isRefresh,omitempty"`
	IsAPIKey     bool                `json:"isAPIKey,omitempty"`
	AccessAction *base.AccessActions `json:"access,omitempty"`
}

// GenerateAccessToken generates access token
func GenerateAccessToken(authClaims *AuthClaims) (string, error) {
	authClaims.IsRefresh = false
	token, err := GenerateToken(authClaims, accessTokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

// GenerateRefreshToken generates refresh token
func GenerateRefreshToken(authClaims *AuthClaims) (string, error) {
	authClaims.IsRefresh = true
	token, err := GenerateToken(authClaims, refreshTokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}
