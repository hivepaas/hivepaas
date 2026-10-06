package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (uc *UC) GetCurrentAuthByJWT(ctx context.Context, jwt string) (*basedto.Auth, error) {
	user, err := uc.GetCurrentUserByJWT(ctx, jwt)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	auth := &basedto.Auth{User: user}
	return auth, uc.verifyAuth(ctx, auth)
}

func (uc *UC) GetCurrentAuthByAPIKey(ctx context.Context, keyID, secret string) (*basedto.Auth, error) {
	user, err := uc.GetCurrentUserByAPIKey(ctx, keyID, secret)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	auth := &basedto.Auth{User: user}
	return auth, uc.verifyAuth(ctx, auth)
}

func (uc *UC) verifyAuth(ctx context.Context, auth *basedto.Auth) error {
	user := auth.User

	// The account may still sign in: not expired, and active
	if err := signInRefusal(user.Entity()); err != nil {
		return err
	}

	// Use must complete MFA requirement
	if user.SecurityOption == base.UserSecurityPassword2FA && user.TotpSecret == "" {
		return hperrors.Wrap(hperrors.ErrUserNotCompleteMFASetup).
			WithMsgLog("user hasn't completed the MFA setup")
	}

	// Update `last_access` timestamp after each period of few minute.
	// NOTE: We can't update `last_access` timestamp every request due to performance reason.
	timeNow := timeutil.NowUTC()
	if user.LastAccess.IsZero() ||
		timeNow.Sub(user.LastAccess) > config.Current().Session.LastAccessUpdatePeriod {
		user.LastAccess = timeNow
		// Just ignore the error if happens (as this is not important action)
		_ = uc.userRepo.Update(ctx, uc.db, user.Entity(), bunex.UpdateColumns("last_access"))
	}

	return nil
}
