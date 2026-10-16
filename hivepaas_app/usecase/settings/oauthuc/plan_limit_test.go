package oauthuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

func oauthSetting(kind base.OAuthKind, status base.SettingStatus) *entity.Setting {
	return &entity.Setting{Type: base.SettingTypeOAuth, Kind: string(kind), Status: status}
}

// Creating, enabling, or changing to a paid kind adds a paid sign-in provider.
// Editing or disabling one that exists does not: it keeps working on every plan.
func TestAddsPaidSignIn(t *testing.T) {
	const (
		google = base.OAuthKindGoogle
		oidc   = base.OAuthKindOpenIDConnect
		github = base.OAuthKindGithub
	)
	active, pending, disabled := base.SettingStatusActive, base.SettingStatusPending, base.SettingStatusDisabled

	tests := []struct {
		name   string
		before *entity.Setting
		after  *entity.Setting
		want   bool
	}{
		{name: "create a paid kind", after: oauthSetting(google, active), want: true},
		{name: "create a code host", after: oauthSetting(github, active)},

		{name: "change a code host to a paid kind",
			before: oauthSetting(github, active), after: oauthSetting(oidc, active), want: true},
		{name: "change a disabled code host to a paid kind",
			before: oauthSetting(github, disabled), after: oauthSetting(google, disabled), want: true},
		{name: "change one paid kind to another",
			before: oauthSetting(google, active), after: oauthSetting(oidc, active)},
		{name: "change a paid kind to a code host",
			before: oauthSetting(google, active), after: oauthSetting(github, active)},

		{name: "enable a paid kind",
			before: oauthSetting(google, disabled), after: oauthSetting(google, active), want: true},
		{name: "activate a pending paid kind",
			before: oauthSetting(oidc, pending), after: oauthSetting(oidc, active), want: true},
		{name: "enable a code host",
			before: oauthSetting(github, disabled), after: oauthSetting(github, active)},

		{name: "edit an active paid kind",
			before: oauthSetting(google, active), after: oauthSetting(google, active)},
		{name: "edit a disabled paid kind",
			before: oauthSetting(google, disabled), after: oauthSetting(google, disabled)},
		{name: "disable a paid kind",
			before: oauthSetting(google, active), after: oauthSetting(google, disabled)},
		{name: "edit a code host",
			before: oauthSetting(github, active), after: oauthSetting(github, active)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, addsPaidSignIn(tc.before, tc.after))
		})
	}
}

// Only a change that adds a paid provider is put to the plan, and its refusal
// reaches the caller.
func TestCheckPaidSignIn(t *testing.T) {
	ctx := context.Background()
	license := &fakeLicenseService{err: hperrors.ErrPlanLimitReached}
	uc := &UC{licenseService: license}

	paid := oauthSetting(base.OAuthKindMicrosoftOnline, base.SettingStatusActive)
	assert.NoError(t, uc.checkPaidSignIn(ctx, database.Tx{}, paid, paid))
	assert.Empty(t, license.checks)

	err := uc.checkPaidSignIn(ctx, database.Tx{}, nil, paid)
	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	assert.Equal(t, []*licenseservice.CheckReq{licenseservice.Adding(base.LicenseLimitSignInProviders, 1)},
		license.checks)
}

type fakeLicenseService struct {
	licenseservice.Service
	checks []*licenseservice.CheckReq
	err    error
}

func (f *fakeLicenseService) Check(_ context.Context, _ database.IDB, req *licenseservice.CheckReq) error {
	f.checks = append(f.checks, req)
	return f.err
}
