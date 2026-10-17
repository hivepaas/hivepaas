package sessionuc

import (
	"context"
	"testing"
	"time"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

type mfaTokens struct{ userservice.Service }

func (mfaTokens) GenerateMFAToken(userID string, _ base.MFAType, deviceID string) (string, error) {
	return "mfa:" + userID + ":" + deviceID, nil
}

type passcodeAttempts struct {
	cacherepository.MFAPasscodeRepo
	started []string
}

func (r *passcodeAttempts) Set(_ context.Context, userID string, _ *cacheentity.MFAPasscode, _ time.Duration) error {
	r.started = append(r.started, userID)
	return nil
}

type trustedDevices struct {
	repository.LoginTrustedDeviceRepo
	devices map[string]*entity.LoginTrustedDevice
}

func (r *trustedDevices) GetByUserAndDevice(_ context.Context, _ database.IDB, userID, deviceID string,
	_ ...bunex.SelectQueryOption) (*entity.LoginTrustedDevice, error) {
	if device, found := r.devices[userID+"/"+deviceID]; found {
		return device, nil
	}
	return nil, hperrors.NewNotFound("Device")
}

// newMFAUCTest is newOAuthUCTest with Jane's authenticator app set up, and the laptop she trusted an hour ago.
func newMFAUCTest(t *testing.T) (*UC, *spyAuditService, *passcodeAttempts) {
	t.Helper()
	cfg := &config.Config{}
	cfg.Session.DeviceTrustedPeriod = 48 * time.Hour
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(nil) })

	uc, audit := newOAuthUCTest(t)
	jane := uc.userRepo.(*userByEmail).users["jane@example.com"]
	jane.TotpSecret = "totp-secret"
	attempts := &passcodeAttempts{}
	uc.userService = mfaTokens{}
	uc.cacheMfaPasscodeRepo = attempts
	uc.loginTrustedDeviceRepo = &trustedDevices{devices: map[string]*entity.LoginTrustedDevice{
		"u-jane/laptop": {UserID: "u-jane", DeviceID: "laptop", UpdatedAt: timeutil.NowUTC().Add(-time.Hour)},
		"u-jane/old":    {UserID: "u-jane", DeviceID: "old", UpdatedAt: timeutil.NowUTC().Add(-72 * time.Hour)},
	}}
	return uc, audit, attempts
}

func janeSignsInWithGoogle(deviceID string) *sessiondto.CreateOAuthSessionReq {
	return &sessiondto.CreateOAuthSessionReq{TrustedDeviceID: deviceID, User: &goth.User{
		Provider: "set-google", Email: "jane@example.com", RawData: map[string]any{"verified_email": true}}}
}

// A provider stands for the password, not for the second factor: a user with an authenticator app owes the
// passcode, and gets the token to sign in with it, not a session.
func TestCreateOAuthSessionAsksForThePasscode(t *testing.T) {
	for _, deviceID := range []string{"", "phone", "old"} {
		uc, audit, attempts := newMFAUCTest(t)

		resp, err := uc.CreateOAuthSession(context.Background(), janeSignsInWithGoogle(deviceID))

		require.NoError(t, err, deviceID)
		assert.Equal(t, "mfa:u-jane:"+deviceID, resp.MFAToken, "the device is trusted once the passcode is right")
		assert.Empty(t, resp.AccessToken, "no session yet")
		assert.Equal(t, []string{"u-jane"}, attempts.started, "the passcode's attempts are counted afresh")
		assert.Empty(t, audit.entries, "nobody signed in yet")
	}
}

func TestPasscodeRequired(t *testing.T) {
	uc, _, _ := newMFAUCTest(t)
	jane := uc.userRepo.(*userByEmail).users["jane@example.com"]

	for deviceID, want := range map[string]bool{"": true, "phone": true, "old": true, "laptop": false} {
		required, err := uc.passcodeRequired(context.Background(), jane, deviceID)
		require.NoError(t, err)
		assert.Equal(t, want, required, deviceID)
	}

	required, err := uc.passcodeRequired(context.Background(), &entity.User{ID: "u-bob"}, "")
	require.NoError(t, err)
	assert.False(t, required, "no authenticator app, no passcode")
}

// One who may not sign in is refused before the passcode is asked for, and the refusal is recorded.
func TestCreateOAuthSessionRefusesADisabledUserBeforeThePasscode(t *testing.T) {
	uc, audit, attempts := newMFAUCTest(t)
	uc.userRepo.(*userByEmail).users["jane@example.com"].Status = base.UserStatusDisabled

	_, err := uc.CreateOAuthSession(context.Background(), janeSignsInWithGoogle(""))

	require.ErrorIs(t, err, hperrors.ErrUserUnavailable)
	assert.Empty(t, attempts.started)
	require.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogResultDenied, audit.entries[0].Result)
	assert.Equal(t, "u-jane", audit.entries[0].ResID)
}
