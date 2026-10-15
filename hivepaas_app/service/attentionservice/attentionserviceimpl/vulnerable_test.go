package attentionserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
)

func vulnApp(id, projectKey string) *entity.App {
	return &entity.App{ID: id, Name: "app " + id, ProjectID: "P1", ProjectEnvID: "E1", Status: base.AppStatusActive,
		Project:    &entity.Project{ID: "P1", Key: projectKey, Name: "Shop"},
		ProjectEnv: &entity.ProjectEnv{ID: "E1", Key: "production", Name: "production"}}
}

// An app whose image carries a fixable CRITICAL is vulnerable; not one whose
// CRITICAL is set aside, nor HivePaaS's own.
func TestVulnerableApps(t *testing.T) {
	now := time.Date(2026, 10, 11, 3, 0, 0, 0, time.UTC)
	critical := func(imageID, vulnID string) *entity.ImageScanFinding {
		return &entity.ImageScanFinding{ImageID: imageID, VulnID: vulnID, PkgName: "openssl", InstalledVersion: "1",
			FixedVersion: "2", Severity: base.VulnSeverityCritical}
	}
	ignore := &entity.Setting{ID: "I1", Type: base.SettingTypeVulnerabilityIgnore, Scope: base.ObjectScopeApp,
		ObjectID: "A2", Status: base.SettingStatusActive}
	ignore.MustSetData(&entity.VulnerabilityIgnore{VulnID: "CVE-2", Reason: "r"})

	items := deriveVulnerable(
		[]*entity.App{vulnApp("A1", "shop"), vulnApp("A2", "shop"), vulnApp("A3", base.HivepaasProjectKey)},
		[]*entity.AppImageScan{
			{AppID: "A1", ImageIDs: []string{"img1"}, ScannedAt: now},
			{AppID: "A2", ImageIDs: []string{"img2"}, ScannedAt: now},
			{AppID: "A3", ImageIDs: []string{"img1"}, ScannedAt: now},
		},
		[]*entity.ImageScanFinding{critical("img1", "CVE-1"), critical("img1", "CVE-3"), critical("img2", "CVE-2")},
		[]*entity.Setting{ignore}, now)

	if assert.Len(t, items, 1) {
		item := items[0]
		assert.Equal(t, attentionservice.KindAppVulnerable, item.Kind)
		assert.Equal(t, attentionservice.SeverityWarning, item.Severity)
		assert.Equal(t, "A1", item.Scope.AppID)
		assert.Equal(t, "production", item.Scope.ProjectEnv)
		assert.Equal(t, "production", item.EnvName)
		assert.Equal(t, 2, item.CriticalFixable)
		assert.Equal(t, now, item.Since)
	}
}

// scanOn is an installation with scanning on.
type scanOn struct{ repository.SettingRepo }

func (scanOn) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	setting := &entity.Setting{Type: base.SettingTypeImageScan}
	setting.MustSetData(&entity.ImageScanSettings{Enabled: true})
	return setting, nil
}

func (scanOn) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return nil, nil, nil
}

// everyScan is every app's scan, and keeps which images' findings were read.
type everyScan struct {
	repository.ImageScanRepo
	read []string
}

func (r *everyScan) ListAppScans(context.Context, database.IDB, ...bunex.SelectQueryOption) (
	[]*entity.AppImageScan, error) {
	return []*entity.AppImageScan{
		{AppID: "A1", ImageIDs: []string{"img1"}},
		{AppID: "unseen", ImageIDs: []string{"img-unseen"}},
	}, nil
}

func (r *everyScan) ListFindings(_ context.Context, _ database.IDB, imageIDs []string,
	_ ...bunex.SelectQueryOption) ([]*entity.ImageScanFinding, error) {
	r.read = imageIDs
	return nil, nil
}

// Home reads the findings of the apps the person sees, not the installation's.
func TestVulnerableItemsReadTheSeenAppsImagesOnly(t *testing.T) {
	scans := &everyScan{}
	s := &service{settingRepo: scanOn{}, imageScanRepo: scans}

	_, err := s.vulnerableItems(context.Background(), nil, []*entity.App{vulnApp("A1", "shop")}, time.Now())

	assert.NoError(t, err)
	assert.Equal(t, []string{"img1"}, scans.read)
}
