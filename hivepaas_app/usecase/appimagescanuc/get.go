package appimagescanuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// GetAppImageScan is what the app's images carry, after the ignores of its
// scopes: its Security tab.
func (uc *UC) GetAppImageScan(
	ctx context.Context,
	_ *basedto.Auth,
	req *appimagescandto.GetAppImageScanReq,
) (*appimagescandto.GetAppImageScanResp, error) {
	app, err := uc.loadApp(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	_, cfg, err := uc.imageScanService.Settings(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := &appimagescandto.AppImageScanResp{
		Enabled:  cfg != nil && cfg.Enabled,
		Summary:  &appimagescandto.SummaryResp{},
		Images:   []*appimagescandto.ImageResp{},
		Findings: []*appimagescandto.FindingResp{},
		Ignored:  []*appimagescandto.IgnoredFindingResp{},
	}

	appScan, err := uc.imageScanRepo.GetAppScan(ctx, uc.db, app.ID)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if appScan != nil {
		if err = uc.fillScans(ctx, app, appScan, data); err != nil {
			return nil, err
		}
	}
	if data.Task, err = uc.pendingScan(ctx, uc.db, app.ID); err != nil {
		return nil, err
	}
	return &appimagescandto.GetAppImageScanResp{Data: data}, nil
}

// fillScans writes the app's images, their findings and those set aside into
// data.
func (uc *UC) fillScans(
	ctx context.Context,
	app *entity.App,
	appScan *entity.AppImageScan,
	data *appimagescandto.AppImageScanResp,
) error {
	data.ScannedAt, data.Error = &appScan.ScannedAt, appScan.Error
	scans, err := uc.imageScanRepo.ListScans(ctx, uc.db, appScan.ImageIDs, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	ignores, err := uc.ignoresOf(ctx, app)
	if err != nil {
		return err
	}

	now := timeutil.NowUTC()
	var total imagescanservice.Summary
	var findings []*entity.ImageScanFinding
	for _, imageID := range appScan.ImageIDs {
		for _, scan := range scans {
			if scan.ImageID != imageID {
				continue
			}
			summary := imagescanservice.Summarize(scan.Findings, ignores, now)
			total.Merge(summary)
			data.Images = append(data.Images, appimagescandto.TransformImage(scan, summary))
			findings = append(findings, scan.Findings...)
		}
	}
	appimagescandto.SortFindings(findings)
	for _, f := range findings {
		if ignore := imagescanservice.Applied(f, ignores, now); ignore != nil {
			data.Ignored = append(data.Ignored, &appimagescandto.IgnoredFindingResp{
				FindingResp: appimagescandto.TransformFinding(f),
				Ignore:      appimagescandto.TransformIgnoreRef(ignore),
			})
			continue
		}
		data.Findings = append(data.Findings, appimagescandto.TransformFinding(f))
	}
	data.Summary = appimagescandto.TransformSummary(total)
	return nil
}

// ignoresOf are the ignores that apply to the app: its own and those its
// scopes have it inherit.
func (uc *UC) ignoresOf(ctx context.Context, app *entity.App) ([]*imagescanservice.Ignore, error) {
	settings, _, err := uc.settingRepo.List(ctx, uc.db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return imagescanservice.IgnoresOf(app, settings), nil
}

// pendingScan is the app's scan not ended yet: the last queued or running.
func (uc *UC) pendingScan(ctx context.Context, db database.IDB, appID string) (
	*appimagescandto.ScanTaskResp, error) {
	tasks, _, err := uc.taskRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeImageScan),
		bunex.SelectWhere("task.object_id = ?", appID),
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
		bunex.SelectOrder("task.created_at DESC"),
		bunex.SelectLimit(1),
		bunex.SelectColumns("id", "status"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return &appimagescandto.ScanTaskResp{ID: tasks[0].ID, Status: tasks[0].Status}, nil
}

func (uc *UC) loadApp(ctx context.Context, projectID, appID string) (*entity.App, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, projectID, appID, true, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	return app, hperrors.Wrap(err)
}
