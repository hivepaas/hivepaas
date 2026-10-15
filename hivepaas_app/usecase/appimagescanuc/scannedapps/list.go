// Package scannedapps lists the scanned apps and what their images carry,
// after the ignores of their scopes, the worst first: every app for System >
// Image Scanning, a project's for its Security tab.
package scannedapps

import (
	"cmp"
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// Repos are what the list reads.
type Repos struct {
	App       repository.AppRepo
	ImageScan repository.ImageScanRepo
	Setting   repository.SettingRepo
}

// Filter narrows the apps listed: to a project, and to the envs and the apps a
// person may read. A nil EnvIDs or AppIDs leaves that unnarrowed.
type Filter struct {
	ProjectID string
	EnvIDs    []string
	AppIDs    []string
}

// List is each scanned app the filter lets through, the worst first.
func List(ctx context.Context, db database.IDB, repos Repos, filter Filter) (
	[]*appimagescandto.ScannedAppResp, error) {
	appScans, err := repos.ImageScan.ListAppScans(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(appScans) == 0 {
		return []*appimagescandto.ScannedAppResp{}, nil
	}
	appIDs := make([]string, 0, len(appScans))
	for _, appScan := range appScans {
		appIDs = append(appIDs, appScan.AppID)
	}
	if filter.AppIDs != nil {
		appIDs = slices.DeleteFunc(appIDs, func(id string) bool { return !slices.Contains(filter.AppIDs, id) })
	}
	if len(appIDs) == 0 || (filter.EnvIDs != nil && len(filter.EnvIDs) == 0) {
		return []*appimagescandto.ScannedAppResp{}, nil
	}
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhereIn("app.id IN (?)", appIDs...),
		bunex.SelectColumns("id", "name", "key", "project_id", "project_env_id", "parent_id"),
		bunex.SelectRelation("Project", bunex.SelectColumns("id", "key", "name")),
		bunex.SelectRelation("ProjectEnv", bunex.SelectColumns("id", "key", "name")),
	}
	if filter.EnvIDs != nil {
		opts = append(opts, bunex.SelectWhereIn("app.project_env_id IN (?)", filter.EnvIDs...))
	}
	apps, _, err := repos.App.List(ctx, db, filter.ProjectID, nil, opts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Only the apps listed have their images read: a project's Security tab
	// reads its own apps' findings, not the whole installation's.
	listed := make(map[string]bool, len(apps))
	for _, app := range apps {
		listed[app.ID] = true
	}
	var imageIDs []string
	for _, appScan := range appScans {
		if listed[appScan.AppID] {
			imageIDs = append(imageIDs, appScan.ImageIDs...)
		}
	}
	scans, err := repos.ImageScan.ListScans(ctx, db, slices.Compact(slices.Sorted(slices.Values(imageIDs))), true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ignoreSettings, _, err := repos.Setting.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return items(apps, appScans, scans, ignoreSettings), nil
}

// Page is the page of items asked for.
func Page(items []*appimagescandto.ScannedAppResp, paging basedto.Paging) (
	*basedto.ListMeta, []*appimagescandto.ScannedAppResp) {
	meta := &basedto.ListMeta{Page: &basedto.PagingMeta{Offset: paging.Offset, Limit: paging.Limit,
		Total: int64(len(items))}}
	start := min(paging.Offset, len(items))
	end := len(items)
	if paging.Limit > 0 {
		end = min(start+paging.Limit, len(items))
	}
	return meta, items[start:end]
}

// items is each app with what its images carry, the worst first: by fixable
// CRITICAL, then fixable HIGH, then by name. HivePaaS's own are not apps of
// any project here.
func items(
	apps []*entity.App,
	appScans []*entity.AppImageScan,
	scans []*entity.ImageScan,
	ignoreSettings []*entity.Setting,
) []*appimagescandto.ScannedAppResp {
	appScanOf := make(map[string]*entity.AppImageScan, len(appScans))
	for _, appScan := range appScans {
		appScanOf[appScan.AppID] = appScan
	}
	scanOf := make(map[string]*entity.ImageScan, len(scans))
	for _, scan := range scans {
		scanOf[scan.ImageID] = scan
	}

	now := timeutil.NowUTC()
	out := make([]*appimagescandto.ScannedAppResp, 0, len(apps))
	totals := make(map[*appimagescandto.ScannedAppResp]imagescanservice.Summary, len(apps))
	for _, app := range apps {
		appScan := appScanOf[app.ID]
		if appScan == nil || app.Project == nil || app.ProjectEnv == nil ||
			app.Project.Key == base.HivepaasProjectKey {
			continue
		}
		ignores := imagescanservice.IgnoresOf(app, ignoreSettings)
		item := &appimagescandto.ScannedAppResp{
			App:        &basedto.NamedObjectResp{ID: app.ID, Name: app.Name},
			Project:    &basedto.NamedObjectResp{ID: app.Project.ID, Name: app.Project.Name},
			ProjectEnv: &appimagescandto.EnvRefResp{ID: app.ProjectEnv.ID, Key: app.ProjectEnv.Key, Name: app.ProjectEnv.Name},
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
		out = append(out, item)
	}
	slices.SortStableFunc(out, func(a, b *appimagescandto.ScannedAppResp) int {
		ta, tb := totals[a], totals[b]
		return cmp.Or(
			-cmp.Compare(ta.Critical.Fixable, tb.Critical.Fixable),
			-cmp.Compare(ta.High.Fixable, tb.High.Fixable),
			cmp.Compare(a.Project.Name, b.Project.Name),
			cmp.Compare(a.ProjectEnv.Name, b.ProjectEnv.Name),
			cmp.Compare(a.App.Name, b.App.Name),
		)
	})
	return out
}
