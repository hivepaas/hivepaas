package attentionserviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
