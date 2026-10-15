package imagescanserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

// resultsKeptFor is how long the results of an image no app runs are kept.
const resultsKeptFor = 30 * timeutil.Day

func (s *service) RunDaily(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.DailyReq,
) (*imagescanservice.DailyResp, error) {
	log := gofn.Coalesce(req.LogStore, tasklog.NewNullStore())
	resp := &imagescanservice.DailyResp{SkipResultNotification: true}
	cfg, err := req.Setting.AsImageScanSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Enabled {
		_ = log.Add(ctx, tasklog.NewOutFrame("Image scanning is switched off: nothing to scan", tasklog.TsNow))
		return resp, nil
	}

	apps, _, err := s.appRepo.List(ctx, db, "", nil, scannedApps()...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ignoreSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The volume is read once for every app and part, as the settings and the
	// ignores are; one that cannot be read leaves each of them not scanned.
	volume, volumeErr := s.loadVolume(ctx, s.db, cfg)
	with := &imageScanReq{cfg: cfg, volume: volume, log: log}

	var news []*notificationservice.ImageScanApp
	failed := 0
	for _, app := range apps {
		if err = ctx.Err(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("--- %s / %s / %s", app.Project.Name,
			app.ProjectEnv.Name, app.Name), tasklog.TsNow))
		scanErr := volumeErr
		if scanErr == nil {
			_, scanErr = s.scanApp(ctx, app, imagescanservice.IgnoresOf(app, ignoreSettings), with)
		}
		if scanErr != nil {
			failed++
			_ = log.Add(ctx, tasklog.NewWarnFrame("Not scanned: "+hperrors.GetErrorDetail(scanErr, ""),
				tasklog.TsNow))
		}
		item, err := s.reportNew(ctx, app, imagescanservice.IgnoresOf(app, ignoreSettings))
		if err != nil {
			return nil, err
		}
		if item != nil {
			news = append(news, item)
		}
	}

	// HivePaaS's own parts: they are fixed by updating HivePaaS, so what they
	// run is shown, never notified.
	parts, partsFailed, err := s.runDailyComponents(ctx, ignoreSettings, with, volumeErr)
	if err != nil {
		return nil, err
	}

	deletedApps, deletedImages, err := s.imageScanRepo.DeleteGone(ctx, s.db,
		timeutil.NowUTC().Add(-resultsKeptFor))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Scanned %d apps, %d not, and %d of HivePaaS's parts, %d "+
		"not; removed the results of %d apps gone and %d images no app runs", len(apps)-failed, failed,
		parts-partsFailed, partsFailed, deletedApps, deletedImages), tasklog.TsNow))
	if total := len(apps) + parts; total > 0 && failed+partsFailed == total {
		// Nothing scanned is the scanner's failure, not the apps': the job's own
		// notification says so.
		return &imagescanservice.DailyResp{}, hperrors.Wrap(hperrors.ErrImageScanFailed).
			WithExtraDetail("No image could be scanned.")
	}

	if len(news) > 0 {
		_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("%d apps run new fixable CRITICAL or HIGH vulnerabilities",
			len(news)), tasklog.TsNow))
		if err = s.notifyNew(ctx, db, req, cfg, news); err != nil {
			_ = log.Add(ctx, tasklog.NewWarnFrame("Could not notify: "+err.Error(), tasklog.TsNow))
		}
	}
	return resp, nil
}

// runDailyComponents scans HivePaaS's parts for the daily run. A volume that
// cannot be read fails them alone: the apps had their own say.
func (s *service) runDailyComponents(ctx context.Context, ignoreSettings []*entity.Setting, with *imageScanReq,
	volumeErr error) (scanned, failed int, err error) {
	if volumeErr != nil {
		_ = with.log.Add(ctx, tasklog.NewWarnFrame("HivePaaS's parts not scanned: "+
			hperrors.GetErrorDetail(volumeErr, ""), tasklog.TsNow))
		return 0, 0, nil
	}
	return s.scanComponents(ctx, ignoreSettings, with)
}

// scannedApps selects the apps scanned every day: those with a service, of
// every project but HivePaaS's own.
func scannedApps() []bunex.SelectQueryOption {
	return []bunex.SelectQueryOption{
		bunex.SelectWhere("app.service_id IS NOT NULL"),
		bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectWhere("project.key != ?", base.HivepaasProjectKey),
		bunex.SelectOrder("project.name", "project_env.name", "app.name"),
	}
}

// reportNew is what the app runs that its last report did not have, fixable
// CRITICAL and HIGH after ignores; nil when nothing is new. What is reported
// is remembered, for the next run to compare with.
func (s *service) reportNew(ctx context.Context, app *entity.App, ignores []*imagescanservice.Ignore) (
	*notificationservice.ImageScanApp, error) {
	appScan, err := s.imageScanRepo.GetAppScan(ctx, s.db, app.ID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	findings, err := s.imageScanRepo.ListFindings(ctx, s.db, appScan.ImageIDs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	now := timeutil.NowUTC()
	keys := imagescanservice.Reportable(findings, ignores, now)
	if slices.Equal(keys, appScan.Reported) {
		return nil, nil
	}

	reported := make(map[string]bool, len(appScan.Reported))
	for _, key := range appScan.Reported {
		reported[key] = true
	}
	item := &notificationservice.ImageScanApp{ProjectName: app.Project.Name, EnvName: app.ProjectEnv.Name,
		AppName: app.Name}
	counted := map[string]bool{}
	for _, f := range findings {
		key := f.Key()
		if reported[key] || counted[key] || !slices.Contains(keys, key) {
			continue
		}
		counted[key] = true
		if f.Severity == base.VulnSeverityCritical {
			item.Critical++
		} else {
			item.High++
		}
	}

	appScan.Reported = keys
	if err = s.imageScanRepo.UpsertAppScan(ctx, s.db, appScan); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if item.Critical+item.High == 0 {
		// Only gone, or set aside: nothing to tell.
		return nil, nil
	}
	item.Link = config.Current().DashboardAppSecurityURL(app.GetObjectScope().GetBaseURLPath())
	return item, nil
}

// notifyNew sends the daily scan's report through the channels its settings
// give.
func (s *service) notifyNew(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.DailyReq,
	cfg *entity.ImageScanSettings,
	news []*notificationservice.ImageScanApp,
) error {
	if cfg.Notification == nil {
		return nil
	}
	scope := entity.NewObjectScopeGlobal()
	notification, err := s.notificationService.GetNotificationForEvent(ctx, db, scope, cfg.Notification, true,
		req.RefObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}
	title := s.notificationService.BuildTitlePrefixForScope(scope) +
		fmt.Sprintf(" %d %s new fixable vulnerabilities", len(news), gofn.If(len(news) == 1, "app runs", "apps run"))
	_, err = s.notificationService.NotifyForTaskResult(ctx, db, &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: true,
		Scope:           scope,
		RefObjects:      req.RefObjects,
		Notification:    notification,
		TemplateName:    notificationservice.TemplateImageScanNotification,
		TemplateData: &notificationservice.TemplateDataImageScan{
			BaseTemplateData: notificationservice.BaseTemplateData{Title: title},
			Apps:             news,
			DashboardLink:    config.Current().DashboardImageScanImagesURL(),
		},
	})
	return hperrors.Wrap(err)
}
