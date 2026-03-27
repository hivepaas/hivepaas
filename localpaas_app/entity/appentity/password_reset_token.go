package appentity

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

type PasswordResetTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userID"`
}
