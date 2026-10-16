package licenseuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/licenseuc/licensedto"
)

// GetLicense is the license in force, with how much of each limit is used.
func (uc *UC) GetLicense(
	ctx context.Context,
	_ *basedto.Auth,
	_ *licensedto.GetLicenseReq,
) (*licensedto.GetLicenseResp, error) {
	state, err := uc.licenseService.Current(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.licenseResp(ctx, state)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &licensedto.GetLicenseResp{Meta: &basedto.Meta{}, Data: data}, nil
}

// licenseResp is the state with the usage now, the portal's pages and the daily
// refresh. Usage takes no lock: it is for showing, and a creation counts again
// under its own.
func (uc *UC) licenseResp(ctx context.Context, state *license.State) (*licensedto.LicenseResp, error) {
	usage, err := uc.licenseService.Usage(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refresh, err := uc.licenseService.RefreshInfo(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := licensedto.TransformLicense(state, usage)
	if resp.InstallationID == "" {
		resp.InstallationID = uc.installationID(ctx)
	}
	resp.Portal = licensedto.TransformLicensePortal(uc.cfg(), resp.InstallationID)
	resp.Refresh = licensedto.TransformLicenseRefresh(refresh)
	return resp, nil
}

// installationID is asked for only when the state has none, which is when no
// key is stored: Current then does not ask docker, as no key is Bee anywhere.
// The page still shows the ID, since a key is bought for it.
//
// A docker that cannot say leaves it empty rather than failing the answer: the
// page that shows it is also the one a key is pasted into, and must load.
func (uc *UC) installationID(ctx context.Context) string {
	id, err := uc.licenseService.InstallationID(ctx)
	if err != nil {
		return ""
	}
	return id
}
