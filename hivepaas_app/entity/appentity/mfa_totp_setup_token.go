package appentity

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"

type MFATotpSetupTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userId"`
	Secret string `json:"secret"`
}
