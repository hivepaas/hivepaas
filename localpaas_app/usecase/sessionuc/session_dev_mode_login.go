package sessionuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

func (uc *UC) DevModeLogin(
	ctx context.Context,
	req *sessiondto.DevModeLoginReq,
) (*sessiondto.DevModeLoginResp, error) {
	user, err := uc.userRepo.GetByID(ctx, uc.db, req.UserID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	sessionData, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{User: user})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sessiondto.DevModeLoginResp{
		Data: &sessiondto.DevModeLoginDataResp{
			AccessToken:     sessionData.AccessToken,
			AccessTokenExp:  sessionData.AccessTokenExp,
			RefreshToken:    sessionData.RefreshToken,
			RefreshTokenExp: sessionData.RefreshTokenExp,
		},
	}, nil
}
