package imagescanuc

import (
	"context"
	"errors"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/scannedapps"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

// ListImageScanComponents is what HivePaaS's own parts run, those the daily
// scan found running, after the global ignores.
func (uc *UC) ListImageScanComponents(
	ctx context.Context,
	_ *basedto.Auth,
	_ *imagescandto.ListImageScanComponentsReq,
) (*imagescandto.ListImageScanComponentsResp, error) {
	appScans, err := uc.imageScanRepo.ListAppScans(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var imageIDs []string
	for _, appScan := range appScans {
		if imagescanservice.IsComponentID(appScan.AppID) {
			imageIDs = append(imageIDs, appScan.ImageIDs...)
		}
	}
	scans, err := uc.imageScanRepo.ListScans(ctx, uc.DB, slices.Compact(slices.Sorted(slices.Values(imageIDs))), true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ignores, err := uc.componentIgnores(ctx)
	if err != nil {
		return nil, err
	}

	now := timeutil.NowUTC()
	data := &imagescandto.ImageScanComponentsDataResp{Components: []*imagescandto.ScannedComponentResp{}}
	for _, component := range imagescanservice.Components {
		idx := slices.IndexFunc(appScans, func(a *entity.AppImageScan) bool { return a.AppID == component.ID() })
		if idx < 0 {
			continue
		}
		appScan := appScans[idx]
		item := &imagescandto.ScannedComponentResp{
			Component: &imagescandto.ComponentRefResp{Key: component.Key, Name: component.Name},
			ScannedAt: appScan.ScannedAt,
			Error:     appScan.Error,
			Images:    []*appimagescandto.ImageResp{},
		}
		var total imagescanservice.Summary
		for _, imageID := range appScan.ImageIDs {
			for _, scan := range scans {
				if scan.ImageID == imageID {
					summary := imagescanservice.Summarize(scan.Findings, ignores, now)
					total.Merge(summary)
					item.Images = append(item.Images, appimagescandto.TransformImage(scan, summary))
				}
			}
		}
		item.Summary = appimagescandto.TransformSummary(total)
		data.Components = append(data.Components, item)
	}
	return &imagescandto.ListImageScanComponentsResp{Data: data}, nil
}

// GetImageScanComponent is what a part of HivePaaS ran when last scanned, its
// findings and those the global ignores set aside.
func (uc *UC) GetImageScanComponent(
	ctx context.Context,
	_ *basedto.Auth,
	req *imagescandto.GetImageScanComponentReq,
) (*imagescandto.GetImageScanComponentResp, error) {
	component, _ := imagescanservice.ComponentByKey(req.Component)
	_, cfg, err := uc.imageScanService.Settings(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	scan := &appimagescandto.AppImageScanResp{
		Enabled:  cfg != nil && cfg.Enabled,
		Summary:  &appimagescandto.SummaryResp{},
		Images:   []*appimagescandto.ImageResp{},
		Findings: []*appimagescandto.FindingResp{},
		Ignored:  []*appimagescandto.IgnoredFindingResp{},
	}
	appScan, err := uc.imageScanRepo.GetAppScan(ctx, uc.DB, component.ID())
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if appScan != nil {
		ignores, err := uc.componentIgnores(ctx)
		if err != nil {
			return nil, err
		}
		if err = scannedapps.FillDetail(ctx, uc.DB, uc.imageScanRepo, appScan, ignores, scan); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &imagescandto.GetImageScanComponentResp{Data: &imagescandto.ComponentImageScanResp{
		Component: &imagescandto.ComponentRefResp{Key: component.Key, Name: component.Name},
		Scan:      scan,
	}}, nil
}

// componentIgnores are the ignores that apply to HivePaaS's parts: the global
// ones.
func (uc *UC) componentIgnores(ctx context.Context) ([]*imagescanservice.Ignore, error) {
	settings, _, err := uc.SettingRepo.List(ctx, uc.DB, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeGlobal),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return imagescanservice.ComponentIgnores(settings), nil
}
