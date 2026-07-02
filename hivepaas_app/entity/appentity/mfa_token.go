package appentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

type MFATokenClaims struct {
	jwtsession.BaseClaims
	Kind            string       `json:"kind"`
	UserID          string       `json:"userId"`
	MFAType         base.MFAType `json:"mfaType"`
	TrustedDeviceID string       `json:"deviceId,omitempty"`
}
