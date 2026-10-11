package imagescanuc

import (
	"cmp"
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

// ListImageScanApps is every scanned app and what its images carry, after the
// ignores of its scopes, the worst first.
func (uc *UC) ListImageScanApps(
	ctx context.Context,
	_ *basedto.Auth,
	req *imagescandto.ListImageScanAppsReq,
) (*imagescandto.ListImageScanAppsResp, error) {
	appScans, err := uc.imageScanRepo.ListAppScans(ctx, uc.DB)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(appScans) == 0 {
		return &imagescandto.ListImageScanAppsResp{Meta: basedto.NewEmptyListMeta(),
			Data: []*imagescandto.ImageScanAppResp{}}, nil
	}
	appIDs := make([]string, 0, len(appScans))
	var imageIDs []string
	for _, appScan := range appScans {
		appIDs = append(appIDs, appScan.AppID)
		imageIDs = append(imageIDs, appScan.ImageIDs...)
	}
	apps, _, err := uc.appRepo.List(ctx, uc.DB, "", nil,
		bunex.SelectWhereIn("app.id IN (?)", appIDs...),
		bunex.SelectColumns("id", "name", "key", "project_id", "project_env_id", "parent_id"),
		bunex.SelectRelation("Project", bunex.SelectColumns("id", "key", "name")),
		bunex.SelectRelation("ProjectEnv", bunex.SelectColumns("id", "name")),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	scans, err := uc.imageScanRepo.ListScans(ctx, uc.DB, slices.Compact(slices.Sorted(slices.Values(imageIDs))), true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ignoreSettings, _, err := uc.SettingRepo.List(ctx, uc.DB, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	items := appItems(apps, appScans, scans, ignoreSettings)
	meta := &basedto.ListMeta{Page: &basedto.PagingMeta{Offset: req.Paging.Offset, Limit: req.Paging.Limit,
		Total: int64(len(items))}}
	start := min(req.Paging.Offset, len(items))
	end := len(items)
	if req.Paging.Limit > 0 {
		end = min(start+req.Paging.Limit, len(items))
	}
	return &imagescandto.ListImageScanAppsResp{Meta: meta, Data: items[start:end]}, nil
}

// appItems is each app with what its images carry, the worst first: by
// fixable CRITICAL, then fixable HIGH, then by name.
func appItems(
	apps []*entity.App,
	appScans []*entity.AppImageScan,
	scans []*entity.ImageScan,
	ignoreSettings []*entity.Setting,
) []*imagescandto.ImageScanAppResp {
	appScanOf := make(map[string]*entity.AppImageScan, len(appScans))
	for _, appScan := range appScans {
		appScanOf[appScan.AppID] = appScan
	}
	scanOf := make(map[string]*entity.ImageScan, len(scans))
	for _, scan := range scans {
		scanOf[scan.ImageID] = scan
	}

	now := timeutil.NowUTC()
	items := make([]*imagescandto.ImageScanAppResp, 0, len(apps))
	totals := make(map[*imagescandto.ImageScanAppResp]imagescanservice.Summary, len(apps))
	for _, app := range apps {
		appScan := appScanOf[app.ID]
		if appScan == nil || app.Project == nil || app.Project.Key == base.HivepaasProjectKey {
			continue
		}
		ignores := imagescanservice.IgnoresOf(app, ignoreSettings)
		item := &imagescandto.ImageScanAppResp{
			App:        &basedto.NamedObjectResp{ID: app.ID, Name: app.Name},
			Project:    &basedto.NamedObjectResp{ID: app.Project.ID, Name: app.Project.Name},
			ProjectEnv: &basedto.NamedObjectResp{ID: app.ProjectEnv.ID, Name: app.ProjectEnv.Name},
			ScannedAt:  appScan.ScannedAt,
			Error:      appScan.Error,
			Images:     []*appimagescandto.ImageResp{},
		}
		var total imagescanservice.Summary
		for _, imageID := range appScan.ImageIDs {
			if scan := scanOf[imageID]; scan != nil {
				summary := imagescanservice.Summarize(scan.Findings, ignores, now)
				total.Merge(summary)
				item.Images = append(item.Images, appimagescandto.TransformImage(scan, summary))
			}
		}
		item.Summary = appimagescandto.TransformSummary(total)
		totals[item] = total
		items = append(items, item)
	}
	slices.SortStableFunc(items, func(a, b *imagescandto.ImageScanAppResp) int {
		ta, tb := totals[a], totals[b]
		return cmp.Or(
			-cmp.Compare(ta.Critical.Fixable, tb.Critical.Fixable),
			-cmp.Compare(ta.High.Fixable, tb.High.Fixable),
			cmp.Compare(a.Project.Name, b.Project.Name),
			cmp.Compare(a.ProjectEnv.Name, b.ProjectEnv.Name),
			cmp.Compare(a.App.Name, b.App.Name),
		)
	})
	return items
}
