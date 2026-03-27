package jwtsession

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

// AuthClaims the claims for authentication
type AuthClaims struct {
	BaseClaims
	UID          string              `json:"uid"`
	UserID       string              `json:"userID"`
	IsRefresh    bool                `json:"isRefresh,omitempty"`
	IsAPIKey     bool                `json:"isAPIKey,omitempty"`
	AccessAction *base.AccessActions `json:"access,omitempty"`
}

// GenerateAccessToken generates access token
func GenerateAccessToken(authClaims *AuthClaims) (string, error) {
	authClaims.IsRefresh = false
	token, err := GenerateToken(authClaims, accessTokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return token, nil
}

// GenerateRefreshToken generates refresh token
func GenerateRefreshToken(authClaims *AuthClaims) (string, error) {
	authClaims.IsRefresh = true
	token, err := GenerateToken(authClaims, refreshTokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return token, nil
}
