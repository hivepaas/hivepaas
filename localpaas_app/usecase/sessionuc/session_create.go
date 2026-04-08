package sessionuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

const (
	uidLen = 16
)

func (uc *UC) createSession(
	ctx context.Context,
	req *sessiondto.BaseCreateSessionReq,
) (resp *sessiondto.BaseCreateSessionResp, err error) {
	authClaims := &jwtsession.AuthClaims{
		UID:          gofn.RandTokenAsHex(uidLen),
		UserID:       req.User.ID,
		IsAPIKey:     req.IsAPIKey,
		AccessAction: req.AccessAction,
	}

	resp = &sessiondto.BaseCreateSessionResp{}
	resp.AccessToken, err = jwtsession.GenerateAccessToken(authClaims)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create access token")
	}
	resp.AccessTokenExp = authClaims.ExpiresAt.Time

	resp.RefreshToken, err = jwtsession.GenerateRefreshToken(authClaims)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create refresh token")
	}
	resp.RefreshTokenExp = authClaims.ExpiresAt.Time

	// Stores the uid in cache, so we can revoke the token later
	err = uc.userTokenRepo.Set(ctx, authClaims.UserID, authClaims.UID, resp.RefreshTokenExp.Sub(timeutil.NowUTC()))
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to store token in cache")
	}

	return resp, nil
}
