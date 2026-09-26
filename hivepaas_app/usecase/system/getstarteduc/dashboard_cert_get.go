package getstarteduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
)

// GetDashboardCert says where the dashboard's certificate stands. The card asks
// only while the installation step is hivepaas/get-started, so the first answer
// that finds it done is also where the step ends - the certificate is the one
// thing the card waits for. That answer still says done, and the card keeps
// telling the admin to reopen the browser until the page is loaded again.
func (uc *UC) GetDashboardCert(
	ctx context.Context,
	auth *basedto.Auth,
	_ *getstarteddto.GetDashboardCertReq,
) (*getstarteddto.GetDashboardCertResp, error) {
	if err := requireAdmin(auth); err != nil {
		return nil, hperrors.Wrap(err)
	}

	item, err := uc.getStartedService.DashboardCert(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if item.Status == getstartedservice.ItemStatusDone {
		if err = uc.getStartedService.Finish(ctx, uc.db); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return &getstarteddto.GetDashboardCertResp{
		Meta: &basedto.Meta{},
		Data: getstarteddto.TransformDashboardCert(item),
	}, nil
}
