package sessionuc

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	// We allow at most 10 attempts of login in the first 2 minutes
	// The duration increase by exponential of 2 after each minute
	maxPasswordFailsInARow       = 10
	passwordCheckDurationEachRow = 2 * time.Minute
	loginAttemptExp              = 4 * time.Hour
)

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
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.New(err)
		}
		if trustedDevice != nil && timeNow.Sub(trustedDevice.UpdatedAt) < config.Current.Session.DeviceTrustedPeriod {
			passcodeRequired = false
		}
	}

	// When passcode is required, builds token for using in the next step
	if passcodeRequired {
		mfaType := base.MFATypeTOTP
		mfaToken, err := uc.userService.GenerateMFAToken(dbUser.ID, mfaType, req.TrustedDeviceID)
		if err != nil {
			return nil, apperrors.New(err)
		}

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
		return nil, apperrors.New(err)
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
		return apperrors.New(err)
	}

	err = uc.userService.VerifyPassword(dbUser, req.Password)
	_ = uc.savePasswordCheckingStatus(ctx, dbUser, attempt, err == nil)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

// allowPasswordLoginAtTheMoment checks if user can do password login at the moment.
// If user made too many login failures, they need to wait for some time before they can try again.
func (uc *UC) allowPasswordLoginAtTheMoment(
	ctx context.Context,
	dbUser *entity.User,
) (*cacheentity.LoginAttempt, error) {
	if dbUser.SecurityOption == base.UserSecurityEnforceSSO {
		return nil, apperrors.New(apperrors.ErrSSORequired)
	}

	attempt, err := uc.cacheLoginAttemptRepo.Get(ctx, dbUser.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, apperrors.New(err)
	}
	if attempt == nil || attempt.Fails < maxPasswordFailsInARow {
		return attempt, nil
	}

	expo := attempt.Fails / maxPasswordFailsInARow
	minWaitingDuration := time.Duration(math.Pow(2, float64(expo))) * passwordCheckDurationEachRow //nolint:mnd
	durationFromFirstFail := timeutil.NowUTC().Sub(attempt.FirstFailAt)
	if durationFromFirstFail > minWaitingDuration {
		return attempt, nil
	}
	waitingDuration := int((minWaitingDuration - durationFromFirstFail).Seconds())
	return nil, apperrors.New(apperrors.ErrTooManyLoginFailures).WithParam("WaitDuration", waitingDuration)
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
				return apperrors.New(err)
			}
		}
		return nil
	}

	// Save failed check count and update the first fail timestamp
	if attempt == nil {
		attempt = &cacheentity.LoginAttempt{}
	}
	attempt.Fails++
	if attempt.FirstFailAt.IsZero() {
		attempt.FirstFailAt = time.Now()
	}
	err := uc.cacheLoginAttemptRepo.Set(ctx, dbUser.ID, attempt, loginAttemptExp)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *UC) wrapSensitiveError(err error) error {
	// Due to security reason, we don't want to send the real error to user for the cases
	// user not found and password mismatched.
	if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrMismatch) ||
		errors.Is(err, apperrors.ErrValueInvalid) {
		// Notes that the `cause` only shows up in dev env, not in production
		return apperrors.New(apperrors.ErrLoginInputInvalid).WithCause(err)
	}
	return apperrors.New(err)
}
