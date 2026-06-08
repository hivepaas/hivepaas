package appentity

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

type ConsoleTokenClaims struct {
	jwtsession.BaseClaims
	Kind     string `json:"kind"`
	UserID   string `json:"userId"`
	TargetID string `json:"targetId"`
}
