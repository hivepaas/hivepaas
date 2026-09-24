package hpappuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// fakeHpAppService serves a fixed release.json and records whether an update
// was started. Methods the tests do not reach are left to the nil interface.
type fakeHpAppService struct {
	hpappservice.Service
	info    *hpappservice.AppReleaseInfo
	updated bool
}

func (s *fakeHpAppService) GetAppReleaseInfo(context.Context) (*hpappservice.AppReleaseInfo, error) {
	return s.info, nil
}

func (s *fakeHpAppService) UpdateSystemVersion(context.Context, database.IDB, *base.ReleaseInfo, bool) error {
	s.updated = true
	return nil
}

func releaseOf(appVersion string, canUpdate bool) *hpappservice.ReleaseInfo {
	return &hpappservice.ReleaseInfo{
		ReleaseInfo: base.ReleaseInfo{AppVersion: appVersion},
		CanUpdate:   canUpdate,
	}
}

func TestUpdateHpApp_RefusesTargetNotNewer(t *testing.T) {
	cases := []struct {
		name   string
		info   *hpappservice.AppReleaseInfo
		target string
		err    error
	}{
		{
			name:   "stable lists an older version",
			info:   &hpappservice.AppReleaseInfo{Stable: releaseOf("v0.0.9", false)},
			target: "v0.0.9",
			err:    hperrors.ErrVersionNotNewer,
		},
		{
			name: "beta lists an older version",
			info: &hpappservice.AppReleaseInfo{
				Stable: releaseOf("v0.2.0", true),
				Beta:   releaseOf("v0.0.9-beta1", false),
			},
			target: "v0.0.9-beta1",
			err:    hperrors.ErrVersionNotNewer,
		},
		{
			name:   "target not listed at all",
			info:   &hpappservice.AppReleaseInfo{Stable: releaseOf("v0.2.0", true)},
			target: "v0.3.0",
			err:    hperrors.ErrUpdateVerMismatched,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeHpAppService{info: tc.info}
			// No db: a refusal has to happen before the transaction is opened.
			uc := New(nil, nil, nil, svc, nil)

			resp, err := uc.UpdateHpApp(context.Background(), &basedto.Auth{},
				&hpappdto.UpdateHpAppReq{TargetVersion: tc.target})

			assert.Nil(t, resp)
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, svc.updated, "no update may be started")
		})
	}
}

// fakeSysUpdate plans what it is told to.
type fakeSysUpdate struct {
	sysupdateservice.Service
	plan *sysupdateservice.UpdatePlan
}

func (s *fakeSysUpdate) PlanUpdate(context.Context, database.IDB, *base.ReleaseInfo) (
	*sysupdateservice.UpdatePlan, error) {
	return s.plan, nil
}

// What the updater would refuse partway through - once the app and workers had
// stopped - is refused before anything is touched.
func TestUpdateHpApp_RefusesWhatTheUpdaterWouldRefuse(t *testing.T) {
	info := &hpappservice.AppReleaseInfo{Stable: releaseOf("v0.2.0", true)}
	cases := []struct {
		name       string
		component  *sysupdateservice.ComponentChange
		skipBackup bool
		err        error
	}{
		{
			name:      "a major the release blocks",
			component: &sysupdateservice.ComponentChange{Key: "db", Change: sysupdateservice.ChangeBlocked},
			err:       hperrors.ErrSystemUpdateBlocked,
		},
		{
			name: "a move made from the backup, with the backup skipped",
			component: &sysupdateservice.ComponentChange{
				Key: "db", Change: sysupdateservice.ChangeMajor, RequiresBackup: true,
			},
			skipBackup: true,
			err:        hperrors.ErrSystemUpdateNeedsBackup,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeHpAppService{info: info}
			plan := &sysupdateservice.UpdatePlan{Components: []*sysupdateservice.ComponentChange{tc.component}}
			// No db: the refusal has to come before the transaction is opened.
			uc := New(nil, nil, nil, svc, &fakeSysUpdate{plan: plan})

			resp, err := uc.UpdateHpApp(context.Background(), &basedto.Auth{},
				&hpappdto.UpdateHpAppReq{TargetVersion: "v0.2.0", SkipBackup: tc.skipBackup})

			assert.Nil(t, resp)
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, svc.updated)
		})
	}
}

// The plan names the target and the channel it is published on.
func TestGetHpAppUpdatePlan(t *testing.T) {
	info := &hpappservice.AppReleaseInfo{
		Current: &hpappservice.CurrentRelease{AppVersion: "v0.1.0", Channel: "stable"},
		Stable:  releaseOf("v0.2.0", true),
		Beta:    releaseOf("v0.3.0-beta1", true),
	}
	info.Beta.NotesURL = "https://example.com/notes"
	plan := &sysupdateservice.UpdatePlan{RequiresBackup: true, Components: []*sysupdateservice.ComponentChange{
		{Key: "db", Change: sysupdateservice.ChangeMajor, RequiresBackup: true},
	}}
	uc := New(nil, nil, nil, &fakeHpAppService{info: info}, &fakeSysUpdate{plan: plan})

	resp, err := uc.GetHpAppUpdatePlan(context.Background(), &basedto.Auth{},
		&hpappdto.GetHpAppUpdatePlanReq{TargetVersion: "v0.3.0-beta1"})

	assert.NoError(t, err)
	assert.Equal(t, "beta", resp.Data.Target.Channel)
	assert.Equal(t, "https://example.com/notes", resp.Data.Target.NotesURL)
	assert.Equal(t, "v0.1.0", resp.Data.Current.AppVersion)
	assert.True(t, resp.Data.RequiresBackup)
	assert.Equal(t, "major", resp.Data.Components[0].Change)

	_, err = uc.GetHpAppUpdatePlan(context.Background(), &basedto.Auth{},
		&hpappdto.GetHpAppUpdatePlanReq{TargetVersion: "v9.9.9"})
	assert.ErrorIs(t, err, hperrors.ErrUpdateVerMismatched)
}
