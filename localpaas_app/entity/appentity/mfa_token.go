package appentity

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

type MFATokenClaims struct {
	jwtsession.BaseClaims
	Kind            string       `json:"kind"`
	UserID          string       `json:"userID"`
	MFAType         base.MFAType `json:"mfaType"`
	TrustedDeviceID string       `json:"deviceID,omitempty"`
}
