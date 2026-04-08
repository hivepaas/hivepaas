package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) RefreshSession(
	ctx context.Context,
	user *basedto.User,
) (resp *sessiondto.RefreshSessionResp, err error) {
	// JWT token must be refresh token
	if !user.AuthClaims.IsRefresh {
		return nil, apperrors.New(apperrors.ErrSessionRefreshTokenRequired)
	}

	sessionData, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{User: user.User})
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create session")
	}

	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.Del(ctx, user.AuthClaims.UserID, user.AuthClaims.UID)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to invalidate old token")
	}

	return &sessiondto.RefreshSessionResp{
		Data: &sessiondto.RefreshSessionDataResp{BaseCreateSessionResp: sessionData},
	}, nil
}
