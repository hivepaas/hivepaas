package attentionserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// vulnerableItems are the apps whose images carry a fixable CRITICAL
// vulnerability no ignore sets aside, while image scanning is on: switched
// off, what was found stays on the apps' screens and leaves this list.
func (s *service) vulnerableItems(ctx context.Context, db database.IDB, apps []*entity.App, now time.Time) (
	[]*attentionservice.Item, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeImageScan, false)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	if cfg, err := setting.AsImageScanSettings(); err != nil || !cfg.Enabled {
		return nil, hperrors.Wrap(err)
	}

	appScans, err := s.imageScanRepo.ListAppScans(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// Only the apps the person sees have their images read.
	seen := make(map[string]bool, len(apps))
	for _, app := range apps {
		seen[app.ID] = true
	}
	var imageIDs []string
	for _, appScan := range appScans {
		if seen[appScan.AppID] {
			imageIDs = append(imageIDs, appScan.ImageIDs...)
		}
	}
	if len(imageIDs) == 0 {
		return nil, nil
	}
	findings, err := s.imageScanRepo.ListFindings(ctx, db, imageIDs,
		bunex.SelectWhere("image_scan_finding.severity = ?", base.VulnSeverityCritical),
		bunex.SelectWhere("image_scan_finding.fixed_version IS NOT NULL"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(findings) == 0 {
		return nil, nil
	}
	ignoreSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return deriveVulnerable(apps, appScans, findings, ignoreSettings, now), nil
}

// deriveVulnerable is an item for each app whose images carry a fixable
// CRITICAL finding no ignore of its scopes sets aside.
func deriveVulnerable(
	apps []*entity.App,
	appScans []*entity.AppImageScan,
	findings []*entity.ImageScanFinding,
	ignoreSettings []*entity.Setting,
	now time.Time,
) []*attentionservice.Item {
	findingsOf := map[string][]*entity.ImageScanFinding{}
	for _, f := range findings {
		findingsOf[f.ImageID] = append(findingsOf[f.ImageID], f)
	}
	appScanOf := make(map[string]*entity.AppImageScan, len(appScans))
	for _, appScan := range appScans {
		appScanOf[appScan.AppID] = appScan
	}

	var items []*attentionservice.Item
	for _, app := range apps {
		appScan := appScanOf[app.ID]
		if appScan == nil || app.Status != base.AppStatusActive || app.Project == nil || app.ProjectEnv == nil ||
			app.Project.Key == base.HivepaasProjectKey {
			continue
		}
		var appFindings []*entity.ImageScanFinding
		for _, imageID := range appScan.ImageIDs {
			appFindings = append(appFindings, findingsOf[imageID]...)
		}
		summary := imagescanservice.Summarize(appFindings, imagescanservice.IgnoresOf(app, ignoreSettings), now)
		if summary.Critical.Fixable == 0 {
			continue
		}
		items = append(items, &attentionservice.Item{
			Kind:     attentionservice.KindAppVulnerable,
			Severity: attentionservice.SeverityWarning,
			Scope: attentionservice.Scope{
				Type:       attentionservice.ScopeApp,
				ProjectID:  app.ProjectID,
				ProjectEnv: app.ProjectEnv.Key,
				AppID:      app.ID,
			},
			Subject:         app.Name,
			ProjectName:     app.Project.Name,
			EnvName:         app.ProjectEnv.Name,
			CriticalFixable: summary.Critical.Fixable,
			Since:           appScan.ScannedAt,
		})
	}
	return items
}
