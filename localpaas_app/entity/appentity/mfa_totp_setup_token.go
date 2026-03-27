package appentity

import "github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"

type MFATotpSetupTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userID"`
	Secret string `json:"secret"`
}
