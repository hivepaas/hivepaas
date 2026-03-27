package appentity

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

type UserInviteTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userID"`
}
