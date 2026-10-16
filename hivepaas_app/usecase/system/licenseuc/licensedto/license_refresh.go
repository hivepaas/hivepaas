package licensedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// RefreshLicenseReq asks the portal now for a newer key than the one stored. It
// has no body: the key asked about is the one stored.
type RefreshLicenseReq struct{}

func NewRefreshLicenseReq() *RefreshLicenseReq {
	return &RefreshLicenseReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RefreshLicenseReq) Validate() hperrors.ValidationErrors {
	return nil
}

type RefreshLicenseResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *RefreshLicenseDataResp `json:"data"`
}

// RefreshLicenseDataResp is what the portal answered and the license in force
// after it, as GET answers it.
//
// The enum field carries no comment of its own: swag would then write it as an
// object holding the enum rather than as the enum.
type RefreshLicenseDataResp struct {
	Result  base.LicenseRefreshResult `json:"result"`
	License *LicenseResp              `json:"license"`
}
