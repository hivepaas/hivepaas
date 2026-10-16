package licenseserviceimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// A key that could never be in force here is refused, and the stored one kept.
func TestInstallRefusesAKeyThatWouldNotBeInForce(t *testing.T) {
	h := newHarness(t)
	expired := payload(func(p *license.Payload) {
		p.IssuedAt, p.ExpiresAt, p.GraceDays = testNow.Add(-60*day), testNow.Add(-20*day), 14
	})

	for name, tc := range map[string]struct {
		key    string
		err    error
		detail string
	}{
		"not a key":              {key: "not a key!", err: hperrors.ErrLicenseInvalid, detail: "neither base64 nor JSON"},
		"signed by nobody known": {key: newSigner(t, "x").keyText(t, payload()), err: hperrors.ErrLicenseInvalid},
		"a development key": {key: h.dev.keyText(t, payload()), err: hperrors.ErrLicenseInvalid,
			detail: "signed by a development key, accepted only in development"},
		"for another installation": {
			key: h.prod.keyText(t, payload(func(p *license.Payload) { p.Installations = []string{otherCluster} })),
			err: hperrors.ErrLicenseOtherInstallation, detail: "This installation's ID is " + thisCluster},
		"past its grace": {key: h.prod.keyText(t, expired), err: hperrors.ErrLicenseExpired},
	} {
		t.Run(name, func(t *testing.T) {
			resp, err := h.Install(context.Background(), tx, &licenseservice.InstallReq{Key: tc.key})

			assert.ErrorIs(t, err, tc.err)
			assert.Nil(t, resp)
			info, _ := hperrors.ParseError(err, translation.LangEn)
			assert.Contains(t, info.Detail, tc.detail)
			assert.NotContains(t, info.Detail, "ops@acme.example")
			assert.NotContains(t, info.Detail, "Acme")
			assert.Empty(t, h.settings.upserted)
		})
	}
}

// A key within its grace still holds, and can be installed.
func TestInstallTakesAKeyInItsGrace(t *testing.T) {
	h := newHarness(t)
	key := h.prod.keyText(t, payload(func(p *license.Payload) {
		p.IssuedAt, p.ExpiresAt, p.GraceDays = testNow.Add(-40*day), testNow.Add(-10*day), 14
	}))

	resp, err := h.Install(context.Background(), tx, &licenseservice.InstallReq{Key: key})

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusGrace, resp.State.Status)
}

// The first key makes the global setting, normalized as a key wrapped by an
// email client would be pasted.
func TestInstallCreatesTheSetting(t *testing.T) {
	h := newHarness(t)
	key := h.prod.keyText(t, payload())
	pasted := "  " + key[:40] + "\n" + key[40:80] + "\r\n\t" + key[80:] + "\n"

	resp, err := h.Install(context.Background(), tx, &licenseservice.InstallReq{Key: pasted})

	assert.NoError(t, err)
	assert.True(t, resp.Created)
	assert.Equal(t, base.LicenseStatusValid, resp.State.Status)
	assert.Equal(t, base.LicensePlanHive, resp.State.Plan)

	assert.Len(t, h.settings.upserted, 1)
	setting := h.settings.upserted[0]
	assert.Equal(t, resp.Setting.ID, setting.ID)
	assert.NotEmpty(t, setting.ID)
	assert.Equal(t, base.ObjectScopeGlobal, setting.Scope)
	assert.Empty(t, setting.ObjectID)
	assert.Equal(t, base.SettingTypeLicense, setting.Type)
	assert.Equal(t, base.SettingStatusActive, setting.Status)
	assert.Equal(t, "License", setting.Name)
	assert.Equal(t, entity.CurrentLicenseSettingsVersion, setting.Version)
	assert.Equal(t, testNow, setting.CreatedAt)
	stored, err := setting.AsLicenseSettings()
	assert.NoError(t, err)
	assert.Equal(t, &entity.LicenseSettings{Key: key}, stored)
	assert.Equal(t, []int{1}, h.settings.lockedTo, "the row is read for update")
}

// Any other key replaces the stored one, a lower plan included, and is not moved
// whatever the one it replaces was.
func TestInstallReplacesTheStoredKey(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.storeKey(t, payload(func(p *license.Payload) { p.Plan = base.LicensePlanColony }))
	h.settings.setting.UpdateVer = 3
	h.settings.setting.MustSetData(&entity.LicenseSettings{
		Key: h.settings.setting.MustAsLicenseSettings().Key, MovedAt: testNow.Add(-day)})
	state, err := h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusMoved, state.Status)

	key := h.prod.keyText(t, payload(func(p *license.Payload) { p.Plan = base.LicensePlanComb }))
	h.now = testNow.Add(time.Minute)
	resp, err := h.Install(ctx, tx, &licenseservice.InstallReq{Key: key})

	assert.NoError(t, err)
	assert.False(t, resp.Created)
	setting := h.settings.setting
	assert.Equal(t, "LICENSE", setting.ID)
	assert.Equal(t, 4, setting.UpdateVer)
	assert.Equal(t, testNow.Add(time.Minute), setting.UpdatedAt)
	assert.Equal(t, &entity.LicenseSettings{Key: key}, setting.MustAsLicenseSettings())

	state, err = h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusValid, state.Status)
	assert.Equal(t, base.LicensePlanComb, state.Plan)
}

// After an install the stored key is verified afresh, not answered from what
// was kept of the one before.
func TestInstallDropsTheVerificationKept(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.storeKey(t, payload())
	_, err := h.Current(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, h.verified)

	_, err = h.Install(ctx, tx, &licenseservice.InstallReq{
		Key: h.prod.keyText(t, payload(func(p *license.Payload) { p.Plan = base.LicensePlanColony }))})

	assert.NoError(t, err)
	assert.Nil(t, h.verified)
}

// A key's text is never echoed back in a refusal, even when it is not a key:
// what was pasted may be anything.
func TestInstallDoesNotEchoWhatWasPasted(t *testing.T) {
	h := newHarness(t)
	pasted := `{"customer":{"email":"ops@acme.example"}}`

	_, err := h.Install(context.Background(), tx, &licenseservice.InstallReq{Key: pasted})

	assert.ErrorIs(t, err, hperrors.ErrLicenseInvalid)
	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.False(t, strings.Contains(info.Detail, "acme"), info.Detail)
}
