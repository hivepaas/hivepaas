package appentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

type PasswordResetTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userId"`
}
