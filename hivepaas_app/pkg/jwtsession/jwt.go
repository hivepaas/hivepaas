package jwtsession

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IBaseClaims interface {
	jwt.Claims

	SetExpirationTime(t time.Time)
	SetNotBefore(t time.Time)
	SetIssuedAt(t time.Time)
	SetAudience(aud jwt.ClaimStrings)
	SetIssuer(iss string)
	SetSubject(sub string)
}

// BaseClaims claims struct which implements jwt.Claims interface and has
// ability to set its fields via methods.
type BaseClaims struct {
	jwt.RegisteredClaims
}

func (c *BaseClaims) SetExpirationTime(t time.Time)    { c.ExpiresAt = jwt.NewNumericDate(t) }
func (c *BaseClaims) SetNotBefore(t time.Time)         { c.NotBefore = jwt.NewNumericDate(t) }
func (c *BaseClaims) SetIssuedAt(t time.Time)          { c.IssuedAt = jwt.NewNumericDate(t) }
func (c *BaseClaims) SetAudience(aud jwt.ClaimStrings) { c.Audience = aud }
func (c *BaseClaims) SetIssuer(iss string)             { c.Issuer = iss }
func (c *BaseClaims) SetSubject(sub string)            { c.Subject = sub }

var _ IBaseClaims = (*BaseClaims)(nil)

// GenerateToken generates a token
func GenerateToken(claims IBaseClaims, expireDuration time.Duration) (string, error) {
	if len(signingKey) == 0 {
		return "", fmt.Errorf("empty signing key: %w", ErrConfigInvalid)
	}

	now := funcNow().UTC()
	claims.SetIssuedAt(now)
	claims.SetExpirationTime(now.Add(expireDuration))

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenStr, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
}

// ParseToken parses a token then verifies it is valid
func ParseToken(tokenStr string, claims IBaseClaims) error {
	if len(signingKey) == 0 {
		return fmt.Errorf("empty signing key: %w", ErrConfigInvalid)
	}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return signingKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return fmt.Errorf("failed to parse token: %w", ErrTokenExpired)
		}
		return fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid || token.Method.Alg() != signingMethod.Alg() {
		return fmt.Errorf("token signing method mismatched: %w", ErrTokenInvalid)
	}
	return nil
}
