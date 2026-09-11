package sessionuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	uidLen = 16
)

func (uc *UC) createSession(
	ctx context.Context,
	req *sessiondto.BaseCreateSessionReq,
) (resp *sessiondto.BaseCreateSessionResp, err error) {
	// Refused rather than recorded under an empty method. Every way into HivePaaS
	// ends here, so this is the one place that can make a login path added later
	// say how it authenticated - and a path that never says goes no further.
	if req.Method == "" {
		return nil, hperrors.NewArgumentInvalid("session login method")
	}

	// If user is demo user, restrict permissions to READ only
	if req.User.IsDemoUser() {
		if req.AccessAction == nil {
			req.AccessAction = &base.AccessActions{}
		}
		req.AccessAction.Reset(true, false, false, false)
	}

	authClaims := &jwtsession.AuthClaims{
		UID:          gofn.RandTokenAsHex(uidLen),
		UserID:       req.User.ID,
		IsAPIKey:     req.IsAPIKey,
		AccessAction: req.AccessAction,
	}

	resp = &sessiondto.BaseCreateSessionResp{}
	resp.AccessToken, err = jwtsession.GenerateAccessToken(authClaims)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create access token")
	}
	resp.AccessTokenExp = authClaims.ExpiresAt.Time

	resp.RefreshToken, err = jwtsession.GenerateRefreshToken(authClaims)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create refresh token")
	}
	resp.RefreshTokenExp = authClaims.ExpiresAt.Time

	// Stores the uid in cache, so we can revoke the token later
	err = uc.userTokenRepo.Set(ctx, authClaims.UserID, authClaims.UID, resp.RefreshTokenExp.Sub(timeutil.NowUTC()))
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to store token in cache")
	}

	// Last, and the login does not happen if it cannot be recorded. Nothing has
	// left this function yet, so refusing here costs the caller another sign-in;
	// the alternative is a live session nobody can account for, which is the one
	// thing this record exists to prevent.
	if err = uc.recordLogin(ctx, req.User, req.Method); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}
