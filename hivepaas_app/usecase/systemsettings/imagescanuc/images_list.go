package imagescanuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/scannedapps"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

// ListImageScanApps is every scanned app and what its images carry, after the
// ignores of its scopes, the worst first.
func (uc *UC) ListImageScanApps(
	ctx context.Context,
	_ *basedto.Auth,
	req *imagescandto.ListImageScanAppsReq,
) (*imagescandto.ListImageScanAppsResp, error) {
	items, err := scannedapps.List(ctx, uc.DB, scannedapps.Repos{App: uc.appRepo, ImageScan: uc.imageScanRepo,
		Setting: uc.SettingRepo}, scannedapps.Filter{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	meta, page := scannedapps.Page(items, req.Paging)
	return &imagescandto.ListImageScanAppsResp{Meta: meta, Data: page}, nil
}
