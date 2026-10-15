package scannedapps

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
)

func app(id, name, projectKey string) *entity.App {
	return &entity.App{ID: id, Name: name, ProjectID: "p-" + projectKey, ProjectEnvID: "p-" + projectKey + "_dev",
		Project:    &entity.Project{ID: "p-" + projectKey, Key: projectKey, Name: projectKey},
		ProjectEnv: &entity.ProjectEnv{ID: "p-" + projectKey + "_dev", Key: "dev", Name: "development"}}
}

func finding(imageID, vulnID string, severity base.VulnSeverity, fixed string) *entity.ImageScanFinding {
	return &entity.ImageScanFinding{ImageID: imageID, VulnID: vulnID, PkgName: "pkg", Severity: severity,
		FixedVersion: fixed}
}

// The worst first: fixable CRITICAL, then fixable HIGH, then by name; an
// ignore of its scope takes a finding out of an app's counts; HivePaaS's own
// project is no project's.
func TestItemsComeTheWorstFirst(t *testing.T) {
	now := time.Now()
	apps := []*entity.App{app("a", "alpha", "shop"), app("b", "beta", "shop"), app("c", "gamma", "shop"),
		app("h", "registry", base.HivepaasProjectKey)}
	appScans := []*entity.AppImageScan{
		{AppID: "a", ImageIDs: []string{"i1"}, ScannedAt: now},
		{AppID: "b", ImageIDs: []string{"i2"}, ScannedAt: now},
		{AppID: "c", ImageIDs: []string{"i3"}, ScannedAt: now},
		{AppID: "h", ImageIDs: []string{"i2"}, ScannedAt: now},
	}
	scans := []*entity.ImageScan{
		{ImageID: "i1", Status: base.ImageScanStatusDone, Findings: []*entity.ImageScanFinding{
			finding("i1", "CVE-1", base.VulnSeverityHigh, "2")}},
		{ImageID: "i2", Status: base.ImageScanStatusDone, Findings: []*entity.ImageScanFinding{
			finding("i2", "CVE-2", base.VulnSeverityCritical, "2")}},
		{ImageID: "i3", Status: base.ImageScanStatusDone, Findings: []*entity.ImageScanFinding{
			finding("i3", "CVE-3", base.VulnSeverityCritical, "2")}},
	}
	ignore := &entity.Setting{Type: base.SettingTypeVulnerabilityIgnore, Status: base.SettingStatusActive,
		Scope: base.ObjectScopeApp, ObjectID: "c"}
	ignore.MustSetData(&entity.VulnerabilityIgnore{VulnID: "CVE-3", Reason: "unused"})

	out := items(apps, appScans, scans, []*entity.Setting{ignore})
	names := make([]string, 0, len(out))
	for _, item := range out {
		names = append(names, item.App.Name)
	}
	assert.Equal(t, []string{"beta", "alpha", "gamma"}, names)
	assert.Equal(t, 1, out[0].Summary.Critical.Fixable)
	assert.Equal(t, 0, out[2].Summary.Critical.Total, "gamma's CRITICAL is ignored")
	assert.Equal(t, "dev", out[0].ProjectEnv.Key)
}

// imageScans answers the app scans it holds.
type imageScans struct {
	repository.ImageScanRepo
	appScans []*entity.AppImageScan
}

func (r *imageScans) ListAppScans(context.Context, database.IDB, ...bunex.SelectQueryOption) (
	[]*entity.AppImageScan, error) {
	return r.appScans, nil
}

// A person who may read none of the scanned apps, or no env, is answered none,
// without asking for the apps.
func TestListAnswersNoneTheFilterLetsThrough(t *testing.T) {
	repos := Repos{ImageScan: &imageScans{appScans: []*entity.AppImageScan{{AppID: "a"}}}}
	for _, filter := range []Filter{{AppIDs: []string{"other"}}, {EnvIDs: []string{}}} {
		out, err := List(context.Background(), nil, repos, filter)
		assert.NoError(t, err)
		assert.Empty(t, out)
	}
}

// A page of the items, and how many there are in all.
func TestPage(t *testing.T) {
	all := items(nil, nil, nil, nil)
	meta, page := Page(all, basedto.Paging{Offset: 5, Limit: 10})
	assert.Empty(t, page)
	assert.Equal(t, int64(0), meta.Page.Total)
}

// installation is every scanned app's scan, and keeps which images' scans
// were read.
type installation struct {
	imageScans
	read []string
}

func (r *installation) ListScans(_ context.Context, _ database.IDB, imageIDs []string, _ bool,
	_ ...bunex.SelectQueryOption) ([]*entity.ImageScan, error) {
	r.read = imageIDs
	return nil, nil
}

// projectApps lists the apps of one project only, as the filter asks.
type projectApps struct {
	repository.AppRepo
	apps []*entity.App
}

func (r projectApps) List(context.Context, database.IDB, string, *basedto.Paging, ...bunex.SelectQueryOption) (
	[]*entity.App, *basedto.PagingMeta, error) {
	return r.apps, nil, nil
}

type noIgnores struct{ repository.SettingRepo }

func (noIgnores) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return nil, nil, nil
}

// A project's Security tab reads its own apps' images, not the installation's:
// another project's app is scanned too, and its image is left unread.
func TestListReadsTheListedAppsImagesOnly(t *testing.T) {
	scans := &installation{imageScans: imageScans{appScans: []*entity.AppImageScan{
		{AppID: "mine", ImageIDs: []string{"img-mine"}},
		{AppID: "theirs", ImageIDs: []string{"img-theirs"}},
	}}}
	repos := Repos{App: projectApps{apps: []*entity.App{app("mine", "mine", "shop")}}, ImageScan: scans,
		Setting: noIgnores{}}

	out, err := List(context.Background(), nil, repos, Filter{ProjectID: "p-shop"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"img-mine"}, scans.read)
	assert.Len(t, out, 1)
}
