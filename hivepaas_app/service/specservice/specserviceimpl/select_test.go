package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func TestSelectProjectsExcludesTheHivePaaSProject(t *testing.T) {
	kept := selectProjects([]*entity.Project{
		{ID: "p1", Key: "project_a"},
		{ID: "p2", Key: base.HivepaasProjectKey},
		{ID: "p3", Key: "project_b"},
	})

	assert.Len(t, kept, 2)
	assert.Equal(t, "project_a", kept[0].Key)
	assert.Equal(t, "project_b", kept[1].Key)
}

func TestSelectProjectsIsOrderedByKey(t *testing.T) {
	kept := selectProjects([]*entity.Project{
		{ID: "p3", Key: "zeta"}, {ID: "p1", Key: "alpha"}, {ID: "p2", Key: "mu"},
	})
	assert.Equal(t, []string{"alpha", "mu", "zeta"},
		[]string{kept[0].Key, kept[1].Key, kept[2].Key})
}

func TestSelectAppsExcludesPreviewApps(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectApps([]*entity.App{
		{ID: "a1", Key: "backend"},
		{ID: "a2", Key: "backend-pr-42", ParentID: "a1"},
		{ID: "a3", Key: "frontend"},
	}, "projects/x/envs/dev", report)

	assert.Len(t, kept, 2)
	assert.Equal(t, "backend", kept[0].Key)
	assert.Equal(t, "frontend", kept[1].Key)
	assert.Equal(t, 1, report.CountBySeverity(specmodel.SeveritySkipped))
	assert.Equal(t, specmodel.CodePreviewAppSkipped, report.Issues[0].Code)
	assert.Contains(t, report.Issues[0].Path, "backend-pr-42")
}

func TestSelectSettingsAppliesThePolicyAndReports(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingTypeSSLCert, Name: "a"},
		{ID: "s2", Type: base.SettingTypeAPIKey, Name: "b"},
		{ID: "s3", Type: base.SettingTypeClusterNetwork, Scope: base.ObjectScopeGlobal, Name: "bridge"},
		{ID: "s4", Type: base.SettingTypeClusterNetwork, Scope: base.ObjectScopeProject, Name: "default"},
	}, "global", report)

	assert.Len(t, kept, 2)
	keptIDs := []string{kept[0].ID, kept[1].ID}
	assert.Contains(t, keptIDs, "s1")
	assert.Contains(t, keptIDs, "s4")
	assert.Equal(t, 2, report.CountBySeverity(specmodel.SeveritySkipped))
}

// A setting type with no registered policy must be refused, not exported.
func TestSelectSettingsRefusesAnUnclassifiedType(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingType("brand-new-type"), Name: "x"},
	}, "global", report)

	assert.Empty(t, kept)
	assert.Equal(t, specmodel.CodeTypeUnclassified, report.Issues[0].Code)
}

// Ordering must not depend on what the database returned, or two exports of
// unchanged data would differ.
func TestSelectSettingsIsOrderedDeterministically(t *testing.T) {
	report := &specmodel.Report{}
	settings := []*entity.Setting{
		{ID: "s3", Type: base.SettingTypeSSHKey, Name: "c"},
		{ID: "s1", Type: base.SettingTypeSSLCert, Name: "a"},
		{ID: "s2", Type: base.SettingTypeSSHKey, Name: "b"},
	}
	forward := selectSettings(settings, "global", report)

	reversed := selectSettings([]*entity.Setting{settings[2], settings[0], settings[1]},
		"global", &specmodel.Report{})

	assert.Equal(t,
		[]string{forward[0].ID, forward[1].ID, forward[2].ID},
		[]string{reversed[0].ID, reversed[1].ID, reversed[2].ID})
}

// The scope path is what makes a report entry actionable: "which of my forty
// apps was this". It is passed in rather than derived, so it is worth checking
// that it reaches the report unchanged.
func TestSelectSettingsPutsTheScopePathInTheReport(t *testing.T) {
	report := &specmodel.Report{}
	selectSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingTypeAPIKey, Name: "deploy-key"},
	}, "projects/project_a/envs/dev/apps/backend", report)

	assert.Len(t, report.Issues, 1)
	assert.Equal(t,
		"projects/project_a/envs/dev/apps/backend/api-key/deploy-key",
		report.Issues[0].Path)
}
