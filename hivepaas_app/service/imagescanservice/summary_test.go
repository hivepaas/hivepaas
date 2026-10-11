package imagescanservice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func ignoreSetting(id string, scope base.ObjectScopeType, objectID string, inheritable bool,
	ignore *entity.VulnerabilityIgnore) *entity.Setting {
	setting := &entity.Setting{ID: id, Type: base.SettingTypeVulnerabilityIgnore, Scope: scope, ObjectID: objectID,
		Status: base.SettingStatusActive, Inheritable: inheritable}
	setting.MustSetData(ignore)
	return setting
}

func expiring(setting *entity.Setting, at time.Time) *entity.Setting {
	setting.ExpireAt = at
	return setting
}

func finding(vulnID, pkg string, severity base.VulnSeverity, fixed string) *entity.ImageScanFinding {
	return &entity.ImageScanFinding{VulnID: vulnID, PkgName: pkg, InstalledVersion: "1", Severity: severity,
		FixedVersion: fixed}
}

var app = &entity.App{ID: "A1", ProjectID: "P1", ProjectEnvID: "E1"}

// An app's ignores are its own and those set above it to be inherited; not
// another app's, nor a disabled one, nor one above it that is not inherited.
func TestIgnoresOfAnApp(t *testing.T) {
	ignore := &entity.VulnerabilityIgnore{VulnID: "CVE-1", Reason: "r"}
	disabled := ignoreSetting("off", base.ObjectScopeApp, "A1", false, ignore)
	disabled.Status = base.SettingStatusDisabled
	settings := []*entity.Setting{
		ignoreSetting("own", base.ObjectScopeApp, "A1", false, ignore),
		ignoreSetting("other", base.ObjectScopeApp, "A2", true, ignore),
		ignoreSetting("env", base.ObjectScopeProjectEnv, "E1", true, ignore),
		ignoreSetting("env-not-inherited", base.ObjectScopeProjectEnv, "E1", false, ignore),
		ignoreSetting("project", base.ObjectScopeProject, "P1", true, ignore),
		ignoreSetting("other-project", base.ObjectScopeProject, "P2", true, ignore),
		ignoreSetting("global", base.ObjectScopeGlobal, "", true, ignore),
		disabled,
	}
	var ids []string
	for _, i := range IgnoresOf(app, settings) {
		ids = append(ids, i.Setting.ID)
	}
	assert.Equal(t, []string{"own", "env", "project", "global"}, ids)
}

// An ignore sets aside its vulnerability - of its package when it names one -
// until it expires; the nearest scope's is the one that applies.
func TestAppliedIgnore(t *testing.T) {
	now := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	ignores := IgnoresOf(app, []*entity.Setting{
		ignoreSetting("project", base.ObjectScopeProject, "P1", true,
			&entity.VulnerabilityIgnore{VulnID: "CVE-1", Reason: "r"}),
		ignoreSetting("own", base.ObjectScopeApp, "A1", false,
			&entity.VulnerabilityIgnore{VulnID: "cve-1", Package: "libxml2", Reason: "r"}),
		expiring(ignoreSetting("expired", base.ObjectScopeApp, "A1", false,
			&entity.VulnerabilityIgnore{VulnID: "CVE-2", Reason: "r"}), now.Add(-time.Hour)),
		expiring(ignoreSetting("until", base.ObjectScopeApp, "A1", false,
			&entity.VulnerabilityIgnore{VulnID: "CVE-3", Reason: "r"}), now.Add(time.Hour)),
	})

	assert.Equal(t, "own", Applied(finding("CVE-1", "libxml2", base.VulnSeverityHigh, ""), ignores, now).Setting.ID)
	assert.Equal(t, "project", Applied(finding("CVE-1", "openssl", base.VulnSeverityHigh, ""), ignores, now).Setting.ID)
	assert.Nil(t, Applied(finding("CVE-2", "openssl", base.VulnSeverityHigh, ""), ignores, now))
	assert.NotNil(t, Applied(finding("CVE-3", "openssl", base.VulnSeverityHigh, ""), ignores, now))
	assert.Nil(t, Applied(finding("CVE-3", "openssl", base.VulnSeverityHigh, ""), ignores, now.Add(2*time.Hour)))
}

// What the daily scan reports: fixable CRITICAL and HIGH that nothing sets
// aside, once each.
func TestReportable(t *testing.T) {
	now := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	ignores := IgnoresOf(app, []*entity.Setting{ignoreSetting("own", base.ObjectScopeApp, "A1", false,
		&entity.VulnerabilityIgnore{VulnID: "CVE-4", Reason: "r"})})
	findings := []*entity.ImageScanFinding{
		finding("CVE-1", "a", base.VulnSeverityCritical, "2"),
		finding("CVE-2", "a", base.VulnSeverityHigh, "2"),
		finding("CVE-3", "a", base.VulnSeverityMedium, "2"),
		finding("CVE-4", "a", base.VulnSeverityCritical, "2"),
		finding("CVE-5", "a", base.VulnSeverityCritical, ""),
		finding("CVE-1", "a", base.VulnSeverityCritical, "2"),
	}
	assert.Equal(t, []string{"CVE-1|a||1", "CVE-2|a||1"}, Reportable(findings, ignores, now))

	summary := Summarize(findings, ignores, now)
	assert.Equal(t, Count{Total: 3, Fixable: 2}, summary.Critical)
	assert.Equal(t, 1, summary.Ignored)
	summary.Merge(summary)
	assert.Equal(t, Count{Total: 6, Fixable: 4}, summary.Critical)
	assert.Equal(t, 2, summary.Ignored)
}
