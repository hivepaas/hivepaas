package appimagescanuc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// GetAppImageSBOM is the packages of an image the app runs, as CycloneDX JSON,
// to download: <app>-<short id>.cdx.json.
func (uc *UC) GetAppImageSBOM(
	ctx context.Context,
	_ *basedto.Auth,
	req *appimagescandto.GetAppImageSBOMReq,
) (*appimagescandto.GetAppImageSBOMResp, error) {
	app, err := uc.loadApp(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	content, err := uc.imageScanService.SBOM(ctx, &imagescanservice.SBOMReq{App: app, ImageID: req.ImageID})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	shortID := strings.TrimPrefix(req.ImageID, "sha256:")
	shortID = shortID[:min(len(shortID), 12)] //nolint:mnd // as Docker shortens an image's ID
	return &appimagescandto.GetAppImageSBOMResp{
		FileName: app.Key + "-" + shortID + ".cdx.json",
		Content:  content,
	}, nil
}
