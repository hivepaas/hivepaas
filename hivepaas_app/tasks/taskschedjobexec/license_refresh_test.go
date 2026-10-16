package taskschedjobexec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// fakeLicenseRefresh is the license service as far as a refresh goes: it
// answers fetched, and applies it as applied.
type fakeLicenseRefresh struct {
	licenseservice.Service
	fetchReqs []*licenseservice.FetchRefreshReq
	applied   []*licenseservice.FetchRefreshResp
	fetched   *licenseservice.FetchRefreshResp
	apply     *licenseservice.ApplyRefreshResp
}

func (f *fakeLicenseRefresh) FetchRefresh(_ context.Context, req *licenseservice.FetchRefreshReq) (
	*licenseservice.FetchRefreshResp, error) {
	f.fetchReqs = append(f.fetchReqs, req)
	return f.fetched, nil
}

func (f *fakeLicenseRefresh) ApplyRefresh(_ context.Context, _ database.Tx, fetched *licenseservice.FetchRefreshResp) (
	*licenseservice.ApplyRefreshResp, error) {
	f.applied = append(f.applied, fetched)
	return f.apply, nil
}

func licenseRefreshRun(withTarget bool) *jobRun {
	setting := &entity.Setting{ID: "job", Type: base.SettingTypeSchedJob}
	setting.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeLicenseRefresh,
		TargetSetting: entity.ObjectID{ID: "LICENSE"}})
	refs := entity.NewRefObjects()
	if withTarget {
		refs.RefSettings["LICENSE"] = &entity.Setting{ID: "LICENSE", Type: base.SettingTypeLicense}
	}
	return &jobRun{
		execData:   &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		jobSetting: setting,
		refObjects: refs,
	}
}

// The daily run asks as the daily check - which the config may switch off - and
// applies what it got; it sends no notification of its own.
func TestLicenseRefreshRunFetchesThenApplies(t *testing.T) {
	fetched := &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshNewer, NewKey: "k2"}
	expiresAt := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	service := &fakeLicenseRefresh{fetched: fetched, apply: &licenseservice.ApplyRefreshResp{
		Result: base.LicenseRefreshNewer, Changed: true,
		State: &license.State{Key: &license.Payload{Plan: base.LicensePlanHive, ExpiresAt: expiresAt,
			Customer: license.Customer{Email: "ops@acme.example"}}},
	}}
	run := licenseRefreshRun(true)
	lines := loggedLines(run)

	result, err := (&Executor{licenseService: service}).runJob(context.Background(), database.Tx{}, run)

	assert.NoError(t, err)
	assert.True(t, result.skipNotification)
	assert.Equal(t, []*licenseservice.FetchRefreshReq{{Manual: false}}, service.fetchReqs)
	assert.Equal(t, []*licenseservice.FetchRefreshResp{fetched}, service.applied)
	logged := strings.Join(*lines, "")
	assert.Contains(t, logged, "newer license key, now installed: Hive, valid until 2026-12-01")
	assert.NotContains(t, logged, "ops@acme.example")
}

// A check that failed fails the run, so its one retry asks again.
func TestLicenseRefreshRunFailsWithTheCheck(t *testing.T) {
	service := &fakeLicenseRefresh{
		fetched: &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshFailed},
		apply: &licenseservice.ApplyRefreshResp{Result: base.LicenseRefreshFailed,
			Err: hperrors.Wrap(hperrors.ErrLicensePortalUnavailable)},
	}
	run := licenseRefreshRun(true)
	lines := loggedLines(run)

	_, err := (&Executor{licenseService: service}).runJob(context.Background(), database.Tx{}, run)

	assert.ErrorIs(t, err, hperrors.ErrLicensePortalUnavailable)
	assert.Len(t, service.applied, 1, "the failure is recorded")
	assert.Contains(t, strings.Join(*lines, ""), "The license refresh failed: ERR_LICENSE_PORTAL_UNAVAILABLE")
}

func TestLicenseRefreshRunWithoutItsSetting(t *testing.T) {
	service := &fakeLicenseRefresh{}

	_, err := (&Executor{licenseService: service}).runJob(context.Background(), database.Tx{},
		licenseRefreshRun(false))

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	assert.Empty(t, service.fetchReqs)
}

// Every result has a line of its own.
func TestLicenseRefreshLogNamesEveryResult(t *testing.T) {
	seen := map[string]base.LicenseRefreshResult{}
	for _, result := range base.AllLicenseRefreshResults {
		line := licenseRefreshLog(&licenseservice.ApplyRefreshResp{Result: result})
		assert.NotContains(t, line, "answered", result)
		if other, dup := seen[line]; dup {
			t.Errorf("%s and %s log the same line", result, other)
		}
		seen[line] = result
	}
}
