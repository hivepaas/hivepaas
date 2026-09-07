package sessionuc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/failbackoff"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	// The first 10 wrong passwords in a row are free. From there each further 10
	// doubles the wait, starting at 2 minutes and measured from the most recent
	// failure - so 10 failures cost 4 minutes, 20 cost 8, and so on.
	maxPasswordFailsInARow       = 10
	passwordCheckDurationEachRow = 2 * time.Minute
	loginAttemptExp              = 4 * time.Hour
	mfaPasscodeExp               = 2 * time.Minute

	// dummyHashForTimingAttack is a pre-calculated valid Argon2id hash used to equalize execution time
	// when a non-existent username is checked.
	dummyHashForTimingAttack = "MTIzNDU2Nzg5MA== AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
)

// passwordBackoff makes each wrong password cost more than the last, counted
// against the account rather than the address the guess came from.
var passwordBackoff = failbackoff.Policy{
	MaxFailsInARow: maxPasswordFailsInARow,
	Step:           passwordCheckDurationEachRow,
}

const (
	nextStepMfaInput = "NextMfa"
	nextStepMfaSetup = "NextMfaSetup"
)

func (uc *UC) LoginWithPassword(
	ctx context.Context,
	req *sessiondto.LoginWithPasswordReq,
) (resp *sessiondto.LoginWithPasswordResp, err error) {
	dbUser, err := uc.userRepo.GetByUsernameOrEmail(ctx, uc.db, req.Username, req.Username)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			// Perform dummy password verification to prevent user enumeration via timing attack
			_ = uc.userService.VerifyPassword(&entity.User{Password: dummyHashForTimingAttack}, req.Password)
		}
		return nil, uc.wrapSensitiveError(err)
	}

	err = uc.passwordCheck(ctx, req, dbUser)
	if err != nil {
		return nil, uc.wrapSensitiveError(err)
	}

	passcodeRequired := dbUser.TotpSecret != ""

	// When trusted device is sent
	if passcodeRequired && req.TrustedDeviceID != "" {
		timeNow := timeutil.NowUTC()
		// If the sending trusted device matches the data in DB
		trustedDevice, err := uc.loginTrustedDeviceRepo.GetByUserAndDevice(ctx, uc.db, dbUser.ID, req.TrustedDeviceID)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		if trustedDevice != nil && timeNow.Sub(trustedDevice.UpdatedAt) < config.Current().Session.DeviceTrustedPeriod {
			passcodeRequired = false
		}
	}

	// When passcode is required, builds token for using in the next step
	if passcodeRequired {
		mfaType := base.MFATypeTOTP
		mfaToken, err := uc.userService.GenerateMFAToken(dbUser.ID, mfaType, req.TrustedDeviceID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}

		// Initialize MFA passcode attempts tracker in redis
		_ = uc.cacheMfaPasscodeRepo.Set(ctx, dbUser.ID, &cacheentity.MFAPasscode{Attempts: 0}, mfaPasscodeExp)

		return &sessiondto.LoginWithPasswordResp{
			Data: &sessiondto.LoginWithPasswordDataResp{
				NextStep: nextStepMfaInput,
				MFAType:  mfaType,
				MFAToken: mfaToken,
			},
		}, nil
	}

	// Create a new session as login succeeds
	sessionData, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{User: dbUser})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var nextStep string
	if dbUser.SecurityOption == base.UserSecurityPassword2FA && dbUser.TotpSecret == "" {
		nextStep = nextStepMfaSetup
	}

	return &sessiondto.LoginWithPasswordResp{
		Data: &sessiondto.LoginWithPasswordDataResp{
			Session:  sessionData,
			NextStep: nextStep,
		},
	}, nil
}

func (uc *UC) passwordCheck(
	ctx context.Context,
	req *sessiondto.LoginWithPasswordReq,
	dbUser *entity.User,
) error {
	attempt, err := uc.allowPasswordLoginAtTheMoment(ctx, dbUser)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.userService.VerifyPassword(dbUser, req.Password)
	_ = uc.savePasswordCheckingStatus(ctx, dbUser, attempt, err == nil)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// allowPasswordLoginAtTheMoment checks if user can do password login at the moment.
// If user made too many login failures, they need to wait for some time before they can try again.
//
// The wait is measured from the last failure, so pacing the guesses out no longer
// avoids it - it used to be measured from the first, which meant only a burst was
// ever slowed down. See failbackoff.Policy.Wait.
func (uc *UC) allowPasswordLoginAtTheMoment(
	ctx context.Context,
	dbUser *entity.User,
) (*cacheentity.LoginAttempt, error) {
	if dbUser.SecurityOption == base.UserSecurityEnforceSSO {
		return nil, hperrors.Wrap(hperrors.ErrSSORequired)
	}

	attempt, err := uc.cacheLoginAttemptRepo.Get(ctx, dbUser.ID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	if wait := passwordBackoff.Wait(attempt, timeutil.NowUTC()); wait > 0 {
		return nil, hperrors.Wrap(hperrors.ErrTooManyLoginFailures).
			WithParam("WaitDuration", int(wait.Seconds()))
	}
	return attempt, nil
}

// savePasswordCheckingStatus saves password checking status including the number of failures
// and timestamp of the first fail.
func (uc *UC) savePasswordCheckingStatus(
	ctx context.Context,
	dbUser *entity.User,
	attempt *cacheentity.LoginAttempt,
	success bool,
) error {
	if success {
		if attempt != nil {
			err := uc.cacheLoginAttemptRepo.Del(ctx, dbUser.ID)
			if err != nil {
				return hperrors.Wrap(err)
			}
		}
		return nil
	}

	// Save the failed check count and move the timestamps the backoff reads
	attempt = passwordBackoff.Fail(attempt, timeutil.NowUTC())
	err := uc.cacheLoginAttemptRepo.Set(ctx, dbUser.ID, attempt, loginAttemptExp)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) wrapSensitiveError(err error) error {
	// Due to security reason, we don't want to send the real error to user for the cases
	// user not found and password mismatched.
	if errors.Is(err, hperrors.ErrNotFound) || errors.Is(err, hperrors.ErrMismatch) ||
		errors.Is(err, hperrors.ErrValueInvalid) {
		// Notes that the `cause` only shows up in dev env, not in production
		return hperrors.Wrap(hperrors.ErrLoginInputInvalid).WithCause(err)
	}
	return hperrors.Wrap(err)
}
