package licenseuc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/licenseuc/licensedto"
)

// The GET carries the portal's pages for this installation, and the daily
// refresh with its last check.
func TestGetLicenseCarriesThePortalAndTheRefresh(t *testing.T) {
	uc, licenseService, _ := newTestUC()
	licenseService.current = installed(false).State
	checkedAt := time.Date(2026, 11, 2, 3, 30, 0, 0, time.FixedZone("+07", 7*3600))
	licenseService.refreshInfo = &licenseservice.RefreshInfo{Enabled: false, Refresh: entity.LicenseRefresh{
		LastCheckedAt: checkedAt, LastResult: base.LicenseRefreshFailed, LastError: "The portal could not be reached",
	}}

	resp, err := uc.GetLicense(context.Background(), adminAuth(), licensedto.NewGetLicenseReq())

	assert.NoError(t, err)
	assert.Equal(t, &licensedto.LicensePortalResp{
		URL:        "https://portal.example",
		PricingURL: "https://portal.example/pricing?installationId=" + thisCluster,
		TrialURL:   "https://portal.example/trial?installationId=" + thisCluster,
		AccountURL: "https://portal.example/account/licenses",
	}, resp.Data.Portal)
	utc := checkedAt.UTC()
	assert.Equal(t, &licensedto.LicenseRefreshResp{Enabled: false, LastCheckedAt: &utc,
		LastResult: base.LicenseRefreshFailed, LastError: "The portal could not be reached"}, resp.Data.Refresh)
}

// Never checked, the refresh says only whether it is on.
func TestGetLicenseBeforeAnyRefresh(t *testing.T) {
	uc, licenseService, _ := newTestUC()
	licenseService.current = installed(false).State

	resp, err := uc.GetLicense(context.Background(), adminAuth(), licensedto.NewGetLicenseReq())

	assert.NoError(t, err)
	data, _ := json.Marshal(resp.Data.Refresh)
	assert.JSONEq(t, `{"enabled":true}`, string(data))
}

// A refresh asks as a person does, whether or not the daily check is on, and a
// newer key it installs is recorded as a write of the license setting - by the
// key's license ID, plan, trial and expiry, never the key or the customer.
func TestRefreshLicenseRecordsANewerKey(t *testing.T) {
	uc, licenseService, audit := newTestUC()
	newer := installed(false)
	fetched := &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshNewer, NewKey: keyText}
	licenseService.fetched = fetched
	licenseService.apply = &licenseservice.ApplyRefreshResp{Result: base.LicenseRefreshNewer, Changed: true,
		Setting: newer.Setting, Installed: newer, State: newer.State}
	licenseService.current = newer.State
	auth := adminAuth()

	resp, err := uc.RefreshLicense(context.Background(), auth, licensedto.NewRefreshLicenseReq())

	assert.NoError(t, err)
	assert.Equal(t, []*licenseservice.FetchRefreshReq{{Manual: true}}, licenseService.fetchReqs)
	assert.Equal(t, []*licenseservice.FetchRefreshResp{fetched}, licenseService.applied)
	assert.Equal(t, base.LicenseRefreshNewer, resp.Data.Result)
	assert.Equal(t, base.LicenseStatusTrial, resp.Data.License.Status)
	assert.NotNil(t, resp.Data.License.Portal)
	assert.NotNil(t, resp.Data.License.Refresh)
	if !assert.Len(t, audit.entries, 1) {
		return
	}
	entry := audit.entries[0]
	assert.Equal(t, base.AuditLogTypeSettingUpdate, entry.Type)
	assert.Equal(t, base.AuditLogSourceAPIAction, entry.Source)
	assert.Equal(t, "refresh", entry.Section)
	assert.Same(t, auth, entry.Auth)
	assert.Equal(t, base.ResourceType(base.SettingTypeLicense), entry.ResType)
	assert.Equal(t, "setting-license", entry.ResID)
	assert.Equal(t, "License", entry.ResName)
	assert.JSONEq(t, `{"result":"newer","licenseId":"lic_e2e","plan":"colony","trial":true,`+
		`"expiresAt":"2026-10-31T05:00:00Z"}`, entry.Detail)
	recorded, _ := json.Marshal(entry)
	for _, secret := range []string{keyText, customerName, customerEmail, "acc_1"} {
		assert.NotContains(t, string(recorded), secret)
	}
}

// A move learned of is recorded with when it was learned of.
func TestRefreshLicenseRecordsAMove(t *testing.T) {
	uc, licenseService, audit := newTestUC()
	moved := installed(false)
	moved.State.Status = base.LicenseStatusMoved
	moved.State.MovedAt = time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)
	licenseService.fetched = &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshMoved}
	licenseService.apply = &licenseservice.ApplyRefreshResp{Result: base.LicenseRefreshMoved, Changed: true,
		Setting: moved.Setting, State: moved.State}
	licenseService.current = moved.State

	resp, err := uc.RefreshLicense(context.Background(), adminAuth(), licensedto.NewRefreshLicenseReq())

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseRefreshMoved, resp.Data.Result)
	assert.Equal(t, base.LicenseStatusMoved, resp.Data.License.Status)
	if assert.Len(t, audit.entries, 1) {
		assert.JSONEq(t, `{"result":"moved","licenseId":"lic_e2e","plan":"colony","trial":true,`+
			`"expiresAt":"2026-10-31T05:00:00Z","movedAt":"2026-11-08T08:00:00Z"}`, audit.entries[0].Detail)
	}
}

// A check that changed nothing - the key is the newest, the portal did not
// issue it, there is no key - is answered and not recorded.
func TestRefreshLicenseThatChangedNothingIsNotRecorded(t *testing.T) {
	for _, result := range []base.LicenseRefreshResult{base.LicenseRefreshCurrent, base.LicenseRefreshUnknown,
		base.LicenseRefreshNone} {
		uc, licenseService, audit := newTestUC()
		licenseService.fetched = &licenseservice.FetchRefreshResp{Result: result}
		licenseService.apply = &licenseservice.ApplyRefreshResp{Result: result}
		licenseService.current = installed(false).State

		resp, err := uc.RefreshLicense(context.Background(), adminAuth(), licensedto.NewRefreshLicenseReq())

		assert.NoError(t, err, result)
		assert.Equal(t, result, resp.Data.Result)
		assert.Empty(t, audit.entries, result)
	}
}

// A check that failed is answered as its error, once it is recorded: 503 for a
// portal that could not be reached.
func TestRefreshLicenseThatFailedIsItsError(t *testing.T) {
	uc, licenseService, audit := newTestUC()
	licenseService.fetched = &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshFailed}
	licenseService.apply = &licenseservice.ApplyRefreshResp{Result: base.LicenseRefreshFailed,
		Err: hperrors.Wrap(hperrors.ErrLicensePortalUnavailable)}

	resp, err := uc.RefreshLicense(context.Background(), adminAuth(), licensedto.NewRefreshLicenseReq())

	assert.ErrorIs(t, err, hperrors.ErrLicensePortalUnavailable)
	assert.Equal(t, 503, hperrors.Wrap(err).StatusCode())
	assert.Nil(t, resp)
	assert.Len(t, licenseService.applied, 1, "the failure is recorded")
	assert.Empty(t, audit.entries)
}
