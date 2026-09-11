package sessionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) RefreshSession(
	ctx context.Context,
	user *basedto.User,
) (resp *sessiondto.RefreshSessionResp, err error) {
	// JWT token must be refresh token
	if !user.AuthClaims.IsRefresh {
		return nil, hperrors.Wrap(hperrors.ErrSessionRefreshTokenRequired)
	}

	// The start travels with the session, so its deadline is measured from the
	// login and not from this renewal. A session minted before that claim existed
	// carries nothing here and is dated from now - one more full window, once.
	sessionData, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{
		User:      user.User,
		StartedAt: user.AuthClaims.SessionStartedAt(),
		Method:    auditMethodRefresh,
	})
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create session")
	}

	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.Del(ctx, user.AuthClaims.UserID, user.AuthClaims.UID)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to invalidate old token")
	}

	return &sessiondto.RefreshSessionResp{
		Data: &sessiondto.RefreshSessionDataResp{BaseCreateSessionResp: sessionData},
	}, nil
}
