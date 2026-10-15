package appimagescanuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/scannedapps"
)

// GetProjectImageScan is what the images of a project's apps carry, those of
// the envs and the apps the person may read alone, the worst first.
func (uc *UC) GetProjectImageScan(
	ctx context.Context,
	auth *basedto.Auth,
	req *appimagescandto.GetProjectImageScanReq,
) (*appimagescandto.GetProjectImageScanResp, error) {
	filter := scannedapps.Filter{ProjectID: req.ProjectID}
	// A project read through a grant on some of its envs, or some of its apps,
	// lists theirs alone.
	if allowedAll, envIDs := auth.AllowedProjectEnvIDs(); !allowedAll {
		filter.EnvIDs = append([]string{}, envIDs...)
	}
	if allowedAll, appIDs := auth.AllowedApps(nil); !allowedAll {
		filter.AppIDs = append([]string{}, appIDs...)
	}
	apps, err := scannedapps.List(ctx, uc.db, scannedapps.Repos{App: uc.appRepo, ImageScan: uc.imageScanRepo,
		Setting: uc.settingRepo}, filter)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_, cfg, err := uc.imageScanService.Settings(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appimagescandto.GetProjectImageScanResp{Data: &appimagescandto.ProjectImageScanDataResp{
		Enabled: cfg != nil && cfg.Enabled,
		Apps:    apps,
	}}, nil
}
