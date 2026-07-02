package appentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

type UserInviteTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userId"`
}
