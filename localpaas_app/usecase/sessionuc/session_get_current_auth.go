package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (uc *UC) GetCurrentAuthByJWT(ctx context.Context, jwt string) (*basedto.Auth, error) {
	user, err := uc.GetCurrentUserByJWT(ctx, jwt)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	auth := &basedto.Auth{User: user}
	return auth, uc.verifyAuth(ctx, auth)
}

func (uc *UC) GetCurrentAuthByAPIKey(ctx context.Context, keyID, secret string) (*basedto.Auth, error) {
	user, err := uc.GetCurrentUserByAPIKey(ctx, keyID, secret)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	auth := &basedto.Auth{User: user}
	return auth, uc.verifyAuth(ctx, auth)
}

func (uc *UC) verifyAuth(ctx context.Context, auth *basedto.Auth) error {
	user := auth.User

	// User must have access permission
	if user.IsAccessExpired() {
		return apperrors.New(apperrors.ErrUserUnavailable).
			WithMsgLog("user access expired at: %v", user.AccessExpireAt)
	}

	// Use must complete MFA requirement
	if user.SecurityOption == base.UserSecurityPassword2FA && user.TotpSecret == "" {
		return apperrors.New(apperrors.ErrUserNotCompleteMFASetup).
			WithMsgLog("user hasn't completed the MFA setup")
	}

	// User status is not active
	if user.Status != base.UserStatusActive {
		return apperrors.New(apperrors.ErrUserUnavailable).
			WithMsgLog("user status: %s", user.Status)
	}

	// Update `last_access` timestamp after each period of few minute.
	// NOTE: We can't update `last_access` timestamp every request due to performance reason.
	timeNow := timeutil.NowUTC()
	if user.LastAccess.IsZero() ||
		timeNow.Sub(user.LastAccess) > config.Current.Session.LastAccessUpdatePeriod {
		user.LastAccess = timeNow
		// Just ignore the error if happens (as this is not important action)
		_ = uc.userRepo.Update(ctx, uc.db, user.Entity(), bunex.UpdateColumns("last_access"))
	}

	return nil
}
