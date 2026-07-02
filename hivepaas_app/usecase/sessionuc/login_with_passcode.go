package sessionuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/totp"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	passcodeMaxAttempts = 5
)

func (uc *UC) LoginWithPasscode(
	ctx context.Context,
	req *sessiondto.LoginWithPasscodeReq,
) (resp *sessiondto.LoginWithPasscodeResp, err error) {
	mfaTokenClaims, err := uc.userService.ParseMFAToken(req.MFAToken)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}

	dbUser, err := uc.userRepo.GetByID(ctx, uc.db, mfaTokenClaims.UserID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// Verify passcode TOTP
	if !totp.VerifyPasscode(req.Passcode, dbUser.TotpSecret) {
		passcode, err := uc.cacheMfaPasscodeRepo.Get(ctx, mfaTokenClaims.UserID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return nil, apperrors.New(apperrors.ErrPasscodeMismatched).
					WithMsgLog("need to login with password first")
			}
			return nil, apperrors.New(err)
		}
		if passcode.Attempts >= passcodeMaxAttempts {
			_ = uc.cacheMfaPasscodeRepo.Del(ctx, mfaTokenClaims.UserID)
			return nil, apperrors.New(apperrors.ErrTooManyPasscodeAttempts).
				WithMsgLog("too many passcode attempts: %d", passcode.Attempts)
		}
		// Increase the attempts
		_ = uc.cacheMfaPasscodeRepo.IncrAttempts(ctx, mfaTokenClaims.UserID, passcode)
		return nil, apperrors.New(apperrors.ErrPasscodeMismatched)
	}

	// Removes the passcode in redis
	if err = uc.cacheMfaPasscodeRepo.Del(ctx, mfaTokenClaims.UserID); err != nil {
		return nil, apperrors.New(err)
	}

	// Save trusted device if needs to
	if mfaTokenClaims.TrustedDeviceID != "" {
		timeNow := timeutil.NowUTC()
		trustedDevice := &entity.LoginTrustedDevice{
			UserID:    dbUser.ID,
			DeviceID:  mfaTokenClaims.TrustedDeviceID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		err = uc.loginTrustedDeviceRepo.Upsert(ctx, uc.db, trustedDevice,
			entity.LoginTrustedDeviceUpsertingConflictCols, entity.LoginTrustedDeviceUpsertingUpdateCols)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	// Create a new session as login succeeds
	sessionData, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{User: dbUser})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sessiondto.LoginWithPasscodeResp{
		Data: &sessiondto.LoginWithPasscodeDataResp{
			Session: sessionData,
		},
	}, nil
}
